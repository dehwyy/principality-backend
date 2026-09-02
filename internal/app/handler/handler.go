package handler

import (
	"net/http"
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
}

type PageData struct {
	Title              string
	MinioBaseURL       string
	ActiveTab          string
	Principalities     []PrincipalityView
	Principality       PrincipalityView
	NextPrincipalityID int
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
	requestedPrincipality := strings.Trim(ctx.Param("principalityId"), "/")

	var principality repository.Principality
	var err error

	if requestedPrincipality == "" {
		principality, err = h.Repository.GetFirstPublishedPrincipality()
	} else {
		principalityID, convErr := repository.PrincipalityIDFromString(requestedPrincipality)
		if convErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Неверный идентификатор княжества",
			})
			return
		}
		if ctx.Query("next") == "true" {
			principality, err = h.Repository.GetNextPublishedPrincipality(principalityID)
		} else {
			principality, err = h.Repository.GetPublishedPrincipality(principalityID)
		}
	}
	if err != nil {

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
			NextPrincipalityID: nextPrincipalityID.Int(),
		},
	)
}

func (h *Handler) PrincipalityDraft(ctx *gin.Context) {
	principality, err := h.Repository.GetDraftPrincipality()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Произошла ошибка при получении княжества",
		})
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
		},
	)
}

func (h *Handler) PrincipalityCatalog(ctx *gin.Context) {
	rawMinArea := strings.TrimSpace(ctx.Query("minArea"))
	minArea := repository.SettlementAreaHectaresFromString(strings.ReplaceAll(rawMinArea, ",", "."))

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
		},
	)
}

func principalityView(principality repository.Principality) PrincipalityView {
	return PrincipalityView{
		PrincipalityID:         principality.PrincipalityID.Int(),
		PrincipalityName:       principality.PrincipalityName,
		PrincipalitySummary:    principality.PrincipalitySummary,
		SettlementAreaHectares: principality.SettlementAreaHectares.String(),
		SettlementType:         principality.SettlementType.Title(),
		SettlementTypeCode:     principality.SettlementType.String(),
		ImageKey:               principality.ImageKey,
		VideoKey:               principality.VideoKey,
		LikeCount:              principality.LikeCount(),
	}
}
