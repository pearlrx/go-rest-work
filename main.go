package main

import (
	"fmt"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
	"test-project/database"
	_ "test-project/docs" // Подключаем пакет с автосгенерированными Swagger документами
	"test-project/routers"
)

// @title Test Project API
// @version 1.0
// @description This is a sample server for a user and task management system.

// @host localhost:8000
// @BasePath /

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiHost := os.Getenv("API_HOST")
	apiPort := os.Getenv("API_PORT")

	database.InitDB()
	router := routers.InitRouter()

	swaggerURL := fmt.Sprintf("http://%s:%s/swagger/doc.json")

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(swaggerURL), // Путь к вашему swagger.json файлу
	))

	serverAddress := fmt.Sprintf("%s:%s", apiHost, apiPort)

	log.Fatal(http.ListenAndServe(serverAddress, router))
}
