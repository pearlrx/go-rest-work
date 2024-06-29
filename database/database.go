package database

import (
	"database/sql"
	"log"
	"os"
	"test-project/models"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connStr := os.Getenv("DATABASE_URL")
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
	insertInitialUsers()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			passport_number VARCHAR(20) NOT NULL,
			surname VARCHAR(50) NOT NULL,
			name VARCHAR(50) NOT NULL,
			patronymic VARCHAR(50),
			address VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			name VARCHAR(100),
			hours INTEGER,
			minutes INTEGER,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ,
			start_time TIMESTAMPTZ,
			end_time TIMESTAMPTZ
		);`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func insertInitialUsers() {
	users := []models.User{
		{PassportNumber: "1234 567890", Surname: "Иванов", Name: "Иван", Patronymic: "Иванович", Address: "г. Москва, ул. Ленина, д. 5, кв. 1"},
		{PassportNumber: "2345 678901", Surname: "Петров", Name: "Петр", Patronymic: "Петрович", Address: "г. Санкт-Петербург, Невский пр., д. 10"},
		{PassportNumber: "3456 789012", Surname: "Сидоров", Name: "Сидор", Patronymic: "Сидорович", Address: "г. Казань, ул. Кремлевская, д. 2"},
	}

	stmt, err := DB.Prepare("INSERT INTO users(passport_number, surname, name, patronymic, address, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, user := range users {
		_, err = stmt.Exec(user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address, time.Now(), time.Now())
		if err != nil {
			log.Fatal(err)
		}
	}
}
