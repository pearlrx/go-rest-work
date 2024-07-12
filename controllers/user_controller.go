package controllers

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	logger.Info("User data received: %+v", user)

	// Разделение паспортного номера на серию и номер
	passportParts := strings.Split(user.PassportNumber, " ")
	if len(passportParts) != 2 {
		logger.Warning("Invalid passport format")
		http.Error(w, "Invalid passport format", http.StatusBadRequest)
		return
	}
	passportSerieStr := passportParts[0]
	passportNumberStr := passportParts[1]
	logger.Info("Parsed passport number: series=%s, number=%s", passportSerieStr, passportNumberStr)

	// Проверка корректности серии паспорта
	_, err = strconv.Atoi(passportSerieStr)
	if err != nil {
		logger.Error("Invalid passport series: %v", err)
		http.Error(w, "Invalid passport series", http.StatusBadRequest)
		return
	}

	// Проверка корректности номера паспорта
	_, err = strconv.Atoi(passportNumberStr)
	if err != nil {
		logger.Error("Invalid passport number: %v", err)
		http.Error(w, "Invalid passport number", http.StatusBadRequest)
		return
	}

	db := database.DB

	// SQL запрос для вставки нового пользователя
	query := `
        INSERT INTO users(passport_number, surname, name, patronymic, address, created_at, updated_at)
        VALUES($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
	var userID int
	err = db.QueryRowContext(r.Context(), query, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address, time.Now(), time.Now()).Scan(&userID)
	if err != nil {
		logger.Error("Error executing query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("User created with ID: %d", userID)

	// Устанавливаем ID созданного пользователя
	user.ID = userID

	// Добавление нового пользователя в файл миграции
	addUserToMigrationFile(user)

	// Отправка успешного ответа клиенту
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	logger.Info("Response sent successfully")
}

func addUserToMigrationFile(user models.User) {
	filePath := "migrations/20230707120000_insert_initial_users.sql"
	migrationLine := "INSERT INTO users (passport_number, surname, name, patronymic, address, created_at, updated_at) VALUES " +
		"('" + user.PassportNumber + "', '" + user.Surname + "', '" + user.Name + "', '" + user.Patronymic + "', '" + user.Address + "', NOW(), NOW());\n"

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Warning("Failed to open migration file: %v", err)
		return
	}
	defer file.Close()

	// Запись строки миграции в файл
	if _, err = file.WriteString(migrationLine); err != nil {
		logger.Error("Failed to write to migration file: %v", err)
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

	// Получение ID пользователя из параметров URL
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Warning("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	log.Printf("User ID: %d", id)

	var updatedUser models.User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		logger.Error("Failed to decode request body: %v", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	log.Printf("Updated user data received: %+v", updatedUser)

	if updatedUser.Name == "" || updatedUser.Surname == "" {
		logger.Error("Name and Surname are required fields")
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
		logger.Error("Error executing query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("User with ID %d updated successfully", id)

	updatedUser.ID = id

	// Отправка успешного ответа клиенту
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(updatedUser); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Response sent successfully")
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

	// Получение ID пользователя из параметров URL
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Error("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	logger.Info("User ID: %d", id)

	stmt, err := database.DB.Prepare("DELETE FROM users WHERE id = $1")
	if err != nil {
		logger.Error("Error preparing query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()
	logger.Info("SQL statement prepared successfully")

	// Выполнение SQL запроса
	_, err = stmt.Exec(id)
	if err != nil {
		logger.Error("Error executing query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("User with ID %d deleted successfully", id)

	// Отправка успешного ответа клиенту
	w.WriteHeader(http.StatusOK)
	logger.Info("Response sent successfully")
}

// @Summary Get tasks for a user
// @Description Get tasks based on user ID and optional date range
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param startTime query string false "Start date (YYYY-MM-DD)"
// @Param endTime query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} models.Task
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id}/tasks [get]
func GetUserTasks(w http.ResponseWriter, r *http.Request) {
	logger.Info("GetUserTasks called")

	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Error("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	logger.Info("User ID: %d", userID)

	// Получаем параметры startTime и endTime из URL
	startDateStr := r.URL.Query().Get("startTime")
	endDateStr := r.URL.Query().Get("endTime")

	// Если startTime и endTime не заданы, устанавливаем их как начало и конец времени
	var startDate, endDate time.Time
	if startDateStr == "" || endDateStr == "" {
		startDate = time.Time{}    // или другое начальное значение по вашему выбору
		endDate = time.Now().UTC() // или другое текущее время по вашему выбору
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			logger.Warning("Invalid start_time format: %v", err)
			http.Error(w, "Invalid start_time format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			logger.Warning("Invalid end_time format: %v", err)
			http.Error(w, "Invalid end_time format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	var tasks []models.Task
	query := `
		SELECT id, user_id, name, hours, minutes, created_at, updated_at, start_time, end_time
		FROM tasks
		WHERE user_id = $1`

	// Добавляем условия для времени, если они были указаны
	if !startDate.IsZero() && !endDate.IsZero() {
		query += ` AND start_time >= $2 AND end_time <= $3`
	}

	query += ` ORDER BY hours DESC, minutes DESC`

	// Формируем параметры для выполнения запроса
	paramsList := []interface{}{userID}
	if !startDate.IsZero() && !endDate.IsZero() {
		paramsList = append(paramsList, startDate, endDate)
	}

	rows, err := database.DB.Query(query, paramsList...)
	if err != nil {
		logger.Error("Error executing query: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	logger.Info("SQL query executed successfully")

	// Обработка результатов запроса
	for rows.Next() {
		var task models.Task
		if err = rows.Scan(&task.ID, &task.UserID, &task.Name, &task.Hours, &task.Minutes, &task.CreatedAt, &task.UpdatedAt, &task.StartTime, &task.EndTime); err != nil {
			logger.Error("Error scanning row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	logger.Info("Retrieved %d tasks", len(tasks))

	// Проверка на ошибки при переборе строк
	if err = rows.Err(); err != nil {
		logger.Error("Error iterating over rows: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправка успешного ответа клиенту с данными задач
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(tasks); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Response sent successfully")
}

// @Summary Start a new task
// @Description Create a new task for a user
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.Task true "Task information"
// @Success 201 {object} models.Task
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/start [post]
func StartTask(w http.ResponseWriter, r *http.Request) {
	log.Println("StartTask called")

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		logger.Error("Error decoding request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	logger.Info("Decoded task: %+v", task)

	var userID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE id = $1", task.UserID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warning("User not found: %d", task.UserID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			logger.Error("Error querying user: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	logger.Info("User found: %d", userID)

	task.StartTime = time.Now().UTC()
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	logger.Info("Task timestamps: StartTime=%v, CreatedAt=%v, UpdatedAt=%v", task.StartTime, task.CreatedAt, task.UpdatedAt)

	query := `
		INSERT INTO tasks (user_id, name, created_at, updated_at, start_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err = database.DB.QueryRow(query, task.UserID, task.Name, task.CreatedAt, task.UpdatedAt, task.StartTime).Scan(&task.ID)
	if err != nil {
		logger.Error("Error inserting task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Task created with ID: %d", task.ID)

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(task); err != nil {
		logger.Error("Error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Stop a task
// @Description Stop a task and calculate duration
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.Task true "Task information"
// @Success 200
// @Failure 500 {object} models.ErrorResponse
// @Router /tasks/stop [post]
func StopTask(w http.ResponseWriter, r *http.Request) {
	logger.Info("StopTask called")

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		logger.Error("Error decoding request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	taskID := task.ID
	logger.Info("Task ID: %d", taskID)

	task.EndTime = time.Now().UTC()
	logger.Info("EndTime set to: %v", task.EndTime)

	var startTime time.Time
	err := database.DB.QueryRow("SELECT start_time FROM tasks WHERE id = $1", taskID).Scan(&startTime)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("Task not found: %d", taskID)
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			logger.Error("Error querying task start_time: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	logger.Info("Task start_time: %v", startTime)

	duration := task.EndTime.Sub(startTime)
	task.Hours = int(duration.Hours())
	task.Minutes = int(duration.Minutes()) % 60

	logger.Info("Task duration calculated: %d hours, %d minutes", task.Hours, task.Minutes)

	_, err = database.DB.Exec(`
		UPDATE tasks 
		SET end_time = $1, hours = $2, minutes = $3, updated_at = $4 
		WHERE id = $5`,
		task.EndTime, task.Hours, task.Minutes, time.Now().UTC(), taskID)
	if err != nil {
		logger.Error("Error updating task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("Task updated successfully: %d", taskID)

	w.WriteHeader(http.StatusOK)
}
