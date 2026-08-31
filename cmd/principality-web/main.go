package main

import (
	"log"

	"github.com/dehwyy/principality-backend/internal/app/api"
)

func main() {
	log.Println("Application start!")

	api.StartServer()

	log.Println("Application terminated!")
}
