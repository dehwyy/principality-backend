package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/dehwyy/principality-backend/internal/app/config"
	"github.com/dehwyy/principality-backend/internal/app/dsn"
	"github.com/dehwyy/principality-backend/internal/app/handler"
	"github.com/dehwyy/principality-backend/internal/app/redis"
	"github.com/dehwyy/principality-backend/internal/app/repository"
	"github.com/dehwyy/principality-backend/internal/pkg"

	_ "github.com/dehwyy/principality-backend/docs"
)

// @title Княжества Древней Руси
// @version 1.0
// @description Веб-сервис учёта удельных княжеств Древней Руси
// @contact.name Егор Купрюшин, ИУ5-52Б
// @license.name AS IS (NO WARRANTY)
// @host localhost:8080
// @schemes http
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Значение заголовка: Bearer <access_token из /api/archaeologists/login>
func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()

	rep, errRep := repository.New(&repository.RepositorySettings{
		PostgresDSN:     postgresString,
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
	})
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	redisClient, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}
	defer redisClient.Close()

	hand := handler.NewHandler(rep, conf, redisClient)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
