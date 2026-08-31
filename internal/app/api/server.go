package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/dehwyy/principality-backend/internal/app/config"
	"github.com/dehwyy/principality-backend/internal/app/handler"
	"github.com/dehwyy/principality-backend/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	cfg := config.NewConfig()

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	principalityHandler := handler.NewHandler(repo, cfg)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	r.GET("/principalities/feed/:principalityId", principalityHandler.PrincipalityFeed)
	r.GET("/principalities/draft", principalityHandler.PrincipalityDraft)
	r.GET("/principalities", principalityHandler.PrincipalityCatalog)

	r.Run(cfg.ServerAddr)

	log.Println("Server down")
}
