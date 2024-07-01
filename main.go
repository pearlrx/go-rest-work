package main

import (
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
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
	database.InitDB()
	router := routers.InitRouter()

	// Настройка маршрутов для Swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8000/swagger/doc.json"), // Путь к вашему swagger.json файлу
	))

	log.Fatal(http.ListenAndServe(":8000", router))
}
