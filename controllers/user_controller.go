package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"strconv"
	"strings"
	"test-project/database"
	"test-project/models"
	"time"
)

type People struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Address    string `json:"address"`
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Логика получения пользователей с фильтрацией и пагинацией
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	passportParts := strings.Split(user.PassportNumber, " ")
	if len(passportParts) != 2 {
		http.Error(w, "Invalid passport format", http.StatusBadRequest)
		return
	}
	passportSerieStr := passportParts[0]
	passportNumberStr := passportParts[1]

	_, err = strconv.Atoi(passportSerieStr)
	if err != nil {
		http.Error(w, "Invalid passport series", http.StatusBadRequest)
		return
	}
	_, err = strconv.Atoi(passportNumberStr)
	if err != nil {
		http.Error(w, "Invalid passport number", http.StatusBadRequest)
		return
	}

	db := database.DB

	query := `
        INSERT INTO users(passport_number, surname, name, patronymic, address, created_at, updated_at)
        VALUES($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
	var userID int
	err = db.QueryRowContext(r.Context(), query, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address, time.Now(), time.Now()).Scan(&userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.ID = userID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func SwaggerHandler() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to the generated Swagger JSON
	)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updatedUser models.User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if updatedUser.Name == "" || updatedUser.Surname == "" {
		http.Error(w, "Name and Surname are required fields", http.StatusBadRequest)
		return
	}

	db := database.DB

	query := `
        UPDATE users
        SET name = $1, surname = $2, patronymic = $3, address = $4, updated_at = $5
        WHERE id = $6
    `
	_, err = db.ExecContext(r.Context(), query, updatedUser.Name, updatedUser.Surname, updatedUser.Patronymic, updatedUser.Address, time.Now(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedUser.ID = id

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	stmt, _ := database.DB.Prepare("DELETE FROM users WHERE id = $1")
	_, err := stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetUserTasks(w http.ResponseWriter, r *http.Request) {
	// Логика получения трудозатрат пользователя за период
}

func StartTask(w http.ResponseWriter, r *http.Request) {
	// Логика начала отсчета времени по задаче
}

func StopTask(w http.ResponseWriter, r *http.Request) {
	// Логика завершения отсчета времени по задаче
}
