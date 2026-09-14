package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/dehwyy/principality-backend/internal/app/config"
	"github.com/dehwyy/principality-backend/internal/app/repository"
)

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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/principalities/feed/*principalityId", h.PrincipalityFeed)
	router.GET("/principalities/draft", h.PrincipalityDraft)
	router.GET("/principalities", h.PrincipalityCatalog)
	router.POST("/principalities/draft", h.CreatePrincipalityDraft)
	router.POST("/principalities/:principalityId/publish", h.PublishPrincipality)
	router.POST("/principalities/:principalityId/remove", h.RemovePrincipality)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
