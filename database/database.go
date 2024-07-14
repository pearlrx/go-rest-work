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

	if err = syncSequence(); err != nil {
		log.Fatalf("Error syncing sequence: %v", err)
	}

	logger.Info("Database initialization complete")
}

func runSQLScript(filePath, tableName string) error {
	// Проверка наличия данных в соответствующей таблице
	var count int
	err := DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return fmt.Errorf("could not query database: %v", err)
	}

	if count > 0 {
		logger.Info("Data already exists in the database for table %s. Skipping script execution.", tableName)
		return nil
	}

	// Чтение SQL скрипта из файла
	script, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	// Разделение скрипта на отдельные запросы
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

func syncSequence() error {
	// Определите имя последовательности. Замените 'users_id_seq' на фактическое имя вашей последовательности.
	sequenceName := "users_id_seq"

	// Найдите текущее максимальное значение в таблице.
	var maxID int
	err := DB.QueryRow("SELECT COALESCE(MAX(id), 0) FROM users").Scan(&maxID)
	if err != nil {
		return fmt.Errorf("could not get max ID from users: %v", err)
	}

	// Установите значение последовательности на текущее максимальное значение.
	_, err = DB.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d, false)", sequenceName, maxID))
	if err != nil {
		return fmt.Errorf("could not set sequence value: %v", err)
	}

	logger.Info("Sequence synchronized successfully.")
	return nil
}
