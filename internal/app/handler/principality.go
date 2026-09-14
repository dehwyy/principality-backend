package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dehwyy/principality-backend/internal/app/ds"
	"github.com/dehwyy/principality-backend/internal/app/repository"
)

const (
	foundingDateLayout = "2006-01-02"
	defaultImageURL    = "/static/img/principality_default.jpg"
	defaultVideoURL    = "/static/video/principality_default.mp4"
)

type PrincipalityView struct {
	PrincipalityID       uint
	PrincipalityName     string
	PrincipalitySummary  string
	FoundingDate         string
	FoundingYear         string
	LandCoefficient      string
	LandCoefficientValue string
	ImageURL             string
	VideoURL             string
	LikeCount            int64
}

type PageData struct {
	Title              string
	ActiveTab          string
	DefaultImageURL    string
	DefaultVideoURL    string
	Principalities     []PrincipalityView
	Principality       PrincipalityView
	HasDraft           bool
	NextPrincipalityID uint
	FoundedBefore      string
	TotalCount         int
}

func (h *Handler) PrincipalityFeed(ctx *gin.Context) {
	requestedPrincipality := strings.Trim(ctx.Param("principalityId"), "/")

	var principality *ds.Principality
	var err error

	if requestedPrincipality == "" {
		principality, err = h.Repository.GetFirstPublishedPrincipality()
	} else {
		principalityID, convErr := strconv.Atoi(requestedPrincipality)
		if convErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, convErr)
			return
		}
		if ctx.Query("next") == "true" {
			principality, err = h.Repository.GetNextPublishedPrincipality(uint(principalityID))
		} else {
			principality, err = h.Repository.GetPublishedPrincipality(uint(principalityID))
		}
	}
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.principalityMissing(ctx)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likeCount, err := h.Repository.CountPrincipalityLikesByID(principality.PrincipalityID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	nextPrincipalityID := principality.PrincipalityID
	nextPrincipality, err := h.Repository.GetNextPublishedPrincipality(principality.PrincipalityID)
	if err == nil {
		nextPrincipalityID = nextPrincipality.PrincipalityID
	}

	ctx.HTML(http.StatusOK, "principality_feed.html", PageData{
		Title:              principality.PrincipalityName,
		ActiveTab:          "feed",
		DefaultImageURL:    defaultImageURL,
		DefaultVideoURL:    defaultVideoURL,
		Principality:       h.principalityView(principality, likeCount),
		NextPrincipalityID: nextPrincipalityID,
	})
}

func (h *Handler) PrincipalityDraft(ctx *gin.Context) {
	principality, err := h.Repository.GetDraftPrincipality(ds.CurrentArchaeologist())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	pageData := PageData{
		Title:           "Добавление княжества",
		ActiveTab:       "draft",
		DefaultImageURL: defaultImageURL,
		DefaultVideoURL: defaultVideoURL,
	}
	if principality != nil {
		pageData.HasDraft = true
		pageData.Principality = h.principalityView(principality, 0)
	}

	ctx.HTML(http.StatusOK, "principality_draft.html", pageData)
}

func (h *Handler) PrincipalityCatalog(ctx *gin.Context) {
	rawFoundedBefore := strings.TrimSpace(ctx.Query("foundedBefore"))
	foundedBefore := sql.NullTime{}
	if parsed, err := time.Parse(foundingDateLayout, rawFoundedBefore); err == nil {
		foundedBefore = sql.NullTime{Time: parsed, Valid: true}
	} else {
		rawFoundedBefore = ""
	}

	principalities, err := h.Repository.GetPublishedPrincipalities(foundedBefore)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	likeCounts, err := h.Repository.CountPrincipalityLikes()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	principalityViews := make([]PrincipalityView, 0, len(principalities))
	for i := range principalities {
		principalityViews = append(
			principalityViews,
			h.principalityView(&principalities[i], likeCounts[principalities[i].PrincipalityID]),
		)
	}

	ctx.HTML(http.StatusOK, "principality_catalog.html", PageData{
		Title:           "Княжества Древней Руси",
		ActiveTab:       "catalog",
		DefaultImageURL: defaultImageURL,
		Principalities:  principalityViews,
		FoundedBefore:   rawFoundedBefore,
		TotalCount:      len(principalityViews),
	})
}

func (h *Handler) CreatePrincipalityDraft(ctx *gin.Context) {
	principalityName := strings.TrimSpace(ctx.PostForm("principality_name"))
	if principalityName == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("название княжества не заполнено"))
		return
	}

	existingDraft, err := h.Repository.GetDraftPrincipality(ds.CurrentArchaeologist())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if existingDraft == nil {
		_, err = h.Repository.CreatePrincipalityDraft(principalityName, ds.CurrentArchaeologist())
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.Redirect(http.StatusFound, "/principalities/draft")
}

func (h *Handler) PublishPrincipality(ctx *gin.Context) {
	principalityID, err := strconv.Atoi(ctx.Param("principalityId"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	principalitySummary := strings.TrimSpace(ctx.PostForm("principality_summary"))
	foundingDate, dateErr := time.Parse(foundingDateLayout, strings.TrimSpace(ctx.PostForm("founding_date")))
	landCoefficient, coefficientErr := strconv.ParseFloat(
		strings.ReplaceAll(strings.TrimSpace(ctx.PostForm("land_coefficient")), ",", "."), 64,
	)
	if principalitySummary == "" || dateErr != nil || coefficientErr != nil || landCoefficient <= 0 || landCoefficient > 1 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("для публикации нужны описание, дата основания и коэффициент земли от 0 до 1"))
		return
	}

	err = h.Repository.PublishPrincipality(
		uint(principalityID),
		ds.CurrentArchaeologist(),
		principalitySummary,
		foundingDate,
		landCoefficient,
	)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/principalities")
}

func (h *Handler) RemovePrincipality(ctx *gin.Context) {
	principalityID, err := strconv.Atoi(ctx.Param("principalityId"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.RemovePrincipality(uint(principalityID))
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalityNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/principalities")
}

func (h *Handler) principalityMissing(ctx *gin.Context) {
	ctx.HTML(http.StatusNotFound, "principality_missing.html", PageData{
		Title:     "Княжество не найдено",
		ActiveTab: "feed",
	})
}

func (h *Handler) principalityView(principality *ds.Principality, likeCount int64) PrincipalityView {
	view := PrincipalityView{
		PrincipalityID:      principality.PrincipalityID,
		PrincipalityName:    principality.PrincipalityName,
		PrincipalitySummary: principality.PrincipalitySummary,
		LikeCount:           likeCount,
	}
	if principality.FoundingDate.Valid {
		view.FoundingDate = principality.FoundingDate.Time.Format(foundingDateLayout)
		view.FoundingYear = fmt.Sprintf("%d г.", principality.FoundingDate.Time.Year())
	}
	if principality.LandCoefficient.Valid {
		view.LandCoefficientValue = fmt.Sprintf("%.2f", principality.LandCoefficient.Float64)
		view.LandCoefficient = strings.Replace(view.LandCoefficientValue, ".", ",", 1)
	}
	if principality.ImageKey != "" {
		view.ImageURL = h.Config.MinioBaseURL + "/principality-media/" + principality.ImageKey
	}
	if principality.VideoKey != "" {
		view.VideoURL = h.Config.MinioBaseURL + "/principality-media/" + principality.VideoKey
	}
	return view
}
