package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"test-project/database"
	"test-project/logger"
	"test-project/models"
	"time"
)

type People struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Address    string `json:"address"`
}

// @Summary Get users
// @Description Get a list of users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Results per page"
// @Success 200 {array} models.User
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	logger.Info("GetUsers called")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	logger.Info("Received query params - page:", pageStr, ", limit:", limitStr)

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		logger.Warning("Invalid page value, setting to default 1: %v", err)
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		logger.Warning("Invalid limit value, setting to default 10: %v", err)
		limit = 10
	}

	offset := (page - 1) * limit
	logger.Info("Computed pagination values - page: %d, limit: %d, offset: %d", page, limit, offset)

	query := `
		SELECT id, passport_number, surname, name, patronymic, address, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := database.DB.Query(query, limit, offset)
	if err != nil {
		logger.Error("Error executing query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	logger.Info("Query executed successfully")

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err = rows.Scan(&user.ID, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address, &user.CreatedAt, &user.UpdatedAt); err != nil {
			logger.Error("Error scanning row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error in rows iteration: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Users retrieved: %d", len(users))

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(users); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println("Response sent successfully")
}

// @Summary Create a new user
// @Description Create a new user with passport information
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User information"
// @Success 201 {object} models.User
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	logger.Info("CreateUser called")

	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if newUser.Name == "" || newUser.Surname == "" {
		logger.Error("Name and Surname are required fields")
		http.Error(w, "Name and Surname are required fields", http.StatusBadRequest)
		return
	}

	_, _, err = ValidatePassportNumber(newUser.PassportNumber, w)
	if err != nil {
		return
	}

	db := database.DB // Подключение к базе данных

	var nextFreeID int
	err = db.QueryRow("SELECT get_next_free_user_id()").Scan(&nextFreeID)
	if err != nil {
		logger.Error("Failed to get next free user ID: %v", err)
		http.Error(w, "Failed to get next free user ID", http.StatusInternalServerError)
		return
	}

	newUser.ID = nextFreeID
	newUser.CreatedAt = time.Now().UTC()
	newUser.UpdatedAt = newUser.CreatedAt

	query := `
        INSERT INTO users (id, passport_number, surname, name, patronymic, address, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err = db.ExecContext(r.Context(), query, newUser.ID, newUser.PassportNumber, newUser.Surname, newUser.Name, newUser.Patronymic, newUser.Address, newUser.CreatedAt, newUser.UpdatedAt)
	if err != nil {
		logger.Error("Error inserting user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("User created successfully with ID %d", newUser.ID)

	addUserToMigrationFile(newUser) // Ваш метод для миграций

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(newUser); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Update a user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.User true "Updated user information"
// @Success 200 {object} models.User
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	logger.Info("UpdateUser called")

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Warning("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	logger.Info("User ID: %d", id)

	var updatedUser models.User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	logger.Info("Updated user data received: %+v", updatedUser)

	if updatedUser.Name == "" || updatedUser.Surname == "" {
		logger.Error("Name and Surname are required fields")
		http.Error(w, "Name and Surname are required fields", http.StatusBadRequest)
		return
	}

	_, _, err = ValidatePassportNumber(updatedUser.PassportNumber, w)
	if err != nil {
		return
	}

	db := database.DB

	query := `
        UPDATE users
        SET passport_number = $1, surname = $2, name = $3, patronymic = $4, address = $5, updated_at = $6
        WHERE id = $7
    `
	result, err := db.ExecContext(r.Context(), query, updatedUser.PassportNumber, updatedUser.Surname, updatedUser.Name, updatedUser.Patronymic, updatedUser.Address, time.Now(), id)
	if err != nil {
		logger.Error("Error executing update query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Error getting rows affected: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		logger.Warning("No rows updated for user ID %d", id)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logger.Info("User with ID %d updated successfully", id)

	updateUserInMigrationFile(updatedUser, id)

	updatedUser.ID = id

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(updatedUser); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("Response sent successfully")
}

// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	logger.Info("DeleteUser called")

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Error("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	logger.Info("User ID: %d", id)

	tx, err := database.DB.Begin()
	if err != nil {
		logger.Error("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM tasks WHERE user_id = $1", id)
	if err != nil {
		logger.Error("Error deleting tasks: %v", err)
		tx.Rollback()
		http.Error(w, "Error deleting tasks", http.StatusInternalServerError)
		return
	}
	logger.Info("Tasks for user ID %d deleted successfully", id)

	_, err = tx.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		logger.Error("Error deleting user: %v", err)
		tx.Rollback()
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}
	logger.Info("User with ID %d deleted successfully", id)

	err = tx.Commit()
	if err != nil {
		logger.Error("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	if err = removeUserFromMigrationFile(id); err != nil {
		logger.Error("Error removing user from migration file: %v", err)
		http.Error(w, "Error removing user from migration file", http.StatusInternalServerError)
		return
	}

	if err = removeTaskFromMigrationFile(id); err != nil {
		logger.Error("Error removing task from migration file: %v", err)
		http.Error(w, "Error removing task from migration file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	logger.Info("Response sent successfully")
}
