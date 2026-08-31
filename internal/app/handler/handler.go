package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dehwyy/principality-backend/internal/app/config"
	"github.com/dehwyy/principality-backend/internal/app/repository"
)

type PrincipalityView struct {
	PrincipalityID         int
	PrincipalityName       string
	PrincipalitySummary    string
	SettlementAreaHectares string
	SettlementType         string
	SettlementTypeCode     string
	ImageKey               string
	VideoKey               string
	LikeCount              int
	IsEstimated            bool
	AreaSource             string
}

type PageData struct {
	Title              string
	MinioBaseURL       string
	ActiveTab          string
	Principalities     []PrincipalityView
	Principality       PrincipalityView
	NextPrincipalityID int
	FeedEntryID        int
	MinArea            string
	TotalCount         int
}

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository, c *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Config:     c,
	}
}

func (h *Handler) PrincipalityFeed(ctx *gin.Context) {
	principalityID, err := strconv.Atoi(ctx.Param("principalityId"))
	if err != nil {
		h.principalityMissing(ctx, "Идентификатор княжества должен быть числом")
		return
	}

	var principality repository.Principality
	if ctx.Query("next") == "true" {
		principality, err = h.Repository.GetNextPublishedPrincipality(principalityID)
	} else {
		principality, err = h.Repository.GetPublishedPrincipality(principalityID)
	}
	if err != nil {
		h.principalityMissing(ctx, "Такого княжества нет")
		return
	}

	nextPrincipality, err := h.Repository.GetNextPublishedPrincipality(principality.PrincipalityID)
	nextPrincipalityID := principality.PrincipalityID
	if err == nil {
		nextPrincipalityID = nextPrincipality.PrincipalityID
	}

	ctx.HTML(
		http.StatusOK,
		"principality_feed.html",
		PageData{
			Title:              principality.PrincipalityName,
			MinioBaseURL:       h.Config.MinioBaseURL,
			ActiveTab:          "feed",
			Principality:       principalityView(principality),
			NextPrincipalityID: nextPrincipalityID,
			FeedEntryID:        h.Repository.GetFirstPublishedPrincipalityID(),
		},
	)
}

func (h *Handler) PrincipalityDraft(ctx *gin.Context) {
	principality, err := h.Repository.GetDraftPrincipality()
	if err != nil {
		h.principalityMissing(ctx, "Черновик княжества не найден")
		return
	}

	ctx.HTML(
		http.StatusOK,
		"principality_draft.html",
		PageData{
			Title:        "Добавление княжества",
			MinioBaseURL: h.Config.MinioBaseURL,
			ActiveTab:    "draft",
			Principality: principalityView(principality),
			FeedEntryID:  h.Repository.GetFirstPublishedPrincipalityID(),
		},
	)
}

func (h *Handler) PrincipalityCatalog(ctx *gin.Context) {
	rawMinArea := strings.TrimSpace(ctx.Query("minArea"))
	minArea, err := strconv.ParseFloat(strings.ReplaceAll(rawMinArea, ",", "."), 64)
	if err != nil {
		minArea = 0
	}

	principalities := h.Repository.GetPublishedPrincipalities(minArea)
	principalityViews := make([]PrincipalityView, 0, len(principalities))
	for _, principality := range principalities {
		principalityViews = append(principalityViews, principalityView(principality))
	}

	ctx.HTML(
		http.StatusOK,
		"principality_catalog.html",
		PageData{
			Title:          "Княжества Древней Руси",
			MinioBaseURL:   h.Config.MinioBaseURL,
			ActiveTab:      "catalog",
			Principalities: principalityViews,
			MinArea:        rawMinArea,
			TotalCount:     len(principalityViews),
			FeedEntryID:    h.Repository.GetFirstPublishedPrincipalityID(),
		},
	)
}

func (h *Handler) principalityMissing(ctx *gin.Context, reason string) {
	ctx.HTML(
		http.StatusNotFound,
		"principality_missing.html",
		PageData{
			Title:        reason,
			MinioBaseURL: h.Config.MinioBaseURL,
			FeedEntryID:  h.Repository.GetFirstPublishedPrincipalityID(),
		},
	)
}

func principalityView(principality repository.Principality) PrincipalityView {
	return PrincipalityView{
		PrincipalityID:         principality.PrincipalityID,
		PrincipalityName:       principality.PrincipalityName,
		PrincipalitySummary:    principality.PrincipalitySummary,
		SettlementAreaHectares: formatArea(principality.SettlementAreaHectares),
		SettlementType:         principality.SettlementType.Title(),
		SettlementTypeCode:     string(principality.SettlementType),
		ImageKey:               principality.ImageKey,
		VideoKey:               principality.VideoKey,
		LikeCount:              principality.LikeCount(),
		IsEstimated:            principality.IsEstimated(),
		AreaSource:             principality.AreaSource,
	}
}

func formatArea(areaHectares float64) string {
	if areaHectares == 0 {
		return ""
	}
	return strings.Replace(
		strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.1f", areaHectares), "0"), "."),
		".",
		",",
		1,
	)
}
