package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"test-project/logger"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	connStr := os.Getenv("DATABASE_URL")
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	if err = runSQLScript("migrations/20230707120000_insert_initial_users.sql"); err != nil {
		log.Fatalf("Error running SQL script: %v", err)
	}

	logger.Info("Database initialization complete")
}

func runSQLScript(filePath string) error {
	var count int

	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("could not query database: %v", err)
	}

	if count > 0 {
		logger.Info("Data already exists in the database. Skipping script execution.")
		return nil
	}

	script, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	_, err = DB.Exec(string(script))
	if err != nil {
		return fmt.Errorf("could not execute SQL script: %v", err)
	}

	logger.Info("SQL script executed successfully")
	return nil
}
