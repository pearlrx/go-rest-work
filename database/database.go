package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"test-project/logger"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file: %v", err)
	}

	connStr := os.Getenv("DATABASE_URL")
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Error opening database connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		logger.Fatal("Error pinging database: %v", err)
	}

	if err = runMigrations("migrations"); err != nil {
		logger.Fatal("Error running migrations: %v", err)
	}

	logger.Info("Database initialization complete")
}

func runMigrations(migrationsPath string) error {
	if err := executeMigrationFiles(migrationsPath, ".down.sql"); err != nil {
		return fmt.Errorf("error executing down migrations: %w", err)
	}

	if err := executeMigrationFiles(migrationsPath, ".up.sql"); err != nil {
		return fmt.Errorf("error executing up migrations: %w", err)
	}

	if err := createGetNextFreeUserIDFunction(DB); err != nil {
		return fmt.Errorf("error creating get_next_free_user_id function: %w", err)
	}

	if err := createGetNextFreeTaskIDFunction(DB); err != nil {
		return fmt.Errorf("error creating get_next_free_task_id function: %w", err)
	}

	if err := runSQLScript("migrations/20230707120000_insert_initial_users.sql"); err != nil {
		return fmt.Errorf("error running SQL script for users: %w", err)
	}
	if err := runSQLScript("migrations/20230707120000_insert_initial_tasks.sql"); err != nil {
		return fmt.Errorf("error running SQL script for tasks: %w", err)
	}

	return nil
}

func executeMigrationFiles(migrationsPath string, suffix string) error {
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("could not read migrations directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			filePath := filepath.Join(migrationsPath, file.Name())
			logger.Info("Running migration: %s", filePath)
			if err = runSQLScript(filePath); err != nil {
				return fmt.Errorf("error running SQL script for %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func runSQLScript(filePath string) error {
	script, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %w", err)
	}

	queries := strings.Split(string(script), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		_, err = DB.Exec(query)
		if err != nil {
			return fmt.Errorf("could not execute SQL query: %w", err)
		}
	}

	logger.Info("SQL script executed successfully: %s", filePath)
	return nil
}
