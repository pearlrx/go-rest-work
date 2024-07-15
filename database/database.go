package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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

	if err = runSQLScript("migrations/20230707120000_insert_initial_users.sql", "users"); err != nil {
		log.Fatalf("Error running SQL script: %v", err)
	}
	if err = runSQLScript("migrations/20230707120000_insert_initial_tasks.sql", "tasks"); err != nil {
		log.Fatalf("Error running SQL script: %v", err)
	}
	logger.Info("Database initialization complete")

	if err = updateUserSequence(); err != nil {
		log.Fatalf("Error syncing sequence: %v", err)
	}

	logger.Info("Database initialization complete")
}

func runSQLScript(filePath, tableName string) error {
	var count int
	err := DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return fmt.Errorf("could not query database: %v", err)
	}

	if count > 0 {
		logger.Info("Data already exists in the database for table %s. Skipping script execution.", tableName)
		return nil
	}

	script, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	queries := strings.Split(string(script), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		_, err = DB.Exec(query)
		if err != nil {
			return fmt.Errorf("could not execute SQL query: %v", err)
		}
	}

	logger.Info("SQL script for table %s executed successfully", tableName)
	return nil
}

func updateUserSequence() error {
	_, err := DB.Exec(`SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE(MAX(id), 1)) FROM users`)
	if err != nil {
		return fmt.Errorf("could not update user sequence: %v", err)
	}

	logger.Info("User sequence updated successfully")
	return nil
}
