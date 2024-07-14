package controllers

import (
	"database/sql"
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

	startDateStr := r.URL.Query().Get("startTime")
	endDateStr := r.URL.Query().Get("endTime")

	var startDate, endDate time.Time
	if startDateStr == "" || endDateStr == "" {
		startDate = time.Time{}
		endDate = time.Now().UTC()
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
	logger.Info("StartTask called")

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Warning("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	log.Printf("User ID: %d", id)

	var task models.Task
	if err = json.NewDecoder(r.Body).Decode(&task); err != nil {
		logger.Error("Error decoding request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	logger.Info("Decoded task: %+v", task)

	var userID int
	err = database.DB.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warning("User not found: %d", id)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			logger.Error("Error querying user: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	logger.Info("User found: %d", userID)

	task.UserID = userID
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

	// Добавление SQL миграции в файл
	addTaskToMigrationFile(task)

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

	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		logger.Warning("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	logger.Info("User ID: %d", userID)

	var task models.Task

	task.EndTime = time.Now().UTC()
	logger.Info("EndTime set to: %v", task.EndTime)

	var startTime time.Time
	err = database.DB.QueryRow("SELECT start_time FROM tasks WHERE user_id = $1 AND end_time IS NULL ORDER BY start_time DESC LIMIT 1", userID).Scan(&startTime)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No active task found for user: %d", userID)
			http.Error(w, "No active task found for user", http.StatusNotFound)
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
		WHERE user_id = $5 AND end_time IS NULL`,
		task.EndTime, task.Hours, task.Minutes, time.Now().UTC(), userID)
	if err != nil {
		logger.Error("Error updating task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info("Task for user %d updated successfully", userID)

	w.WriteHeader(http.StatusOK)
}
