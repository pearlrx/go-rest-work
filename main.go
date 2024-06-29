package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"test-project/database"
	"test-project/routers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.InitDB()
	router := routers.InitRouter()
	log.Fatal(http.ListenAndServe(":8000", router))
}
