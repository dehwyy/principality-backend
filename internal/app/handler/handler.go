package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/dehwyy/principality-backend/internal/app/config"
	"github.com/dehwyy/principality-backend/internal/app/redis"
	"github.com/dehwyy/principality-backend/internal/app/repository"
	"github.com/dehwyy/principality-backend/internal/app/role"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
	Redis      *redis.Client
}

func NewHandler(r *repository.Repository, c *config.Config, rc *redis.Client) *Handler {
	return &Handler{
		Repository: r,
		Config:     c,
		Redis:      rc,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/principalities/feed/*principalityId", h.PrincipalityFeed)
	router.GET("/principalities/draft", h.PrincipalityDraft)
	router.GET("/principalities", h.PrincipalityCatalog)
	router.POST("/principalities/draft", h.CreatePrincipalityDraft)
	router.POST("/principalities/:principalityId/publish", h.PublishPrincipality)
	router.POST("/principalities/:principalityId/remove", h.RemovePrincipality)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/api/principalities", h.WithGuestAccess(), h.GetPrincipalitiesAPI)
	router.GET("/api/principalities/feed", h.WithGuestAccess(), h.GetPrincipalityFeedAPI)
	router.GET("/api/principalities/feed/:principalityId", h.WithGuestAccess(), h.GetPrincipalityFeedAPI)
	router.POST("/api/archaeologists", h.RegisterArchaeologistAPI)
	router.POST("/api/archaeologists/login", h.LoginArchaeologistAPI)
	router.POST("/api/archaeologists/logout", h.LogoutArchaeologistAPI)

	archaeologistMethods := router.Group("/api")
	archaeologistMethods.Use(h.WithAuthCheck(role.Archaeologist, role.ChronicleKeeper))
	archaeologistMethods.GET("/principalities/draft", h.GetPrincipalityDraftAPI)
	archaeologistMethods.POST("/principalities", h.AddPrincipalityAPI)
	archaeologistMethods.PUT("/principalities/draft/publish", h.PublishPrincipalityAPI)
	archaeologistMethods.DELETE("/principalities/:principalityId", h.DeletePrincipalityAPI)
	archaeologistMethods.POST("/principalities/:principalityId/like", h.LikePrincipalityAPI)
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
