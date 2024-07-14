package controllers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"test-project/logger"
	"test-project/models"
	"time"
)

func addTaskToMigrationFile(task models.Task) {
	filePath := "migrations/20230707120000_insert_initial_tasks.sql"
	migrationLine := fmt.Sprintf(
		"INSERT INTO tasks (id, user_id, name, created_at, updated_at, start_time) VALUES (%d, %d, '%s', '%s', '%s', '%s');\n",
		task.ID, task.UserID, task.Name, task.CreatedAt.Format(time.RFC3339), task.UpdatedAt.Format(time.RFC3339), task.StartTime.Format(time.RFC3339),
	)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Warning("Failed to open migration file: %v", err)
		return
	}
	defer file.Close()

	if _, err = file.WriteString(migrationLine); err != nil {
		logger.Error("Failed to write to migration file: %v", err)
	} else {
		logger.Info("Successfully added migration line for task %d", task.ID)
	}
}

func addUserToMigrationFile(user models.User) {
	filePath := "migrations/20230707120000_insert_initial_users.sql"
	migrationLine := fmt.Sprintf(
		"INSERT INTO users (id, passport_number, surname, name, patronymic, address, created_at, updated_at) VALUES (%d, '%s', '%s', '%s', '%s', '%s', NOW(), NOW());\n",
		user.ID, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address,
	)

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

func removeUserFromMigrationFile(userID int) error {
	filePath := "migrations/20230707120000_insert_initial_users.sql"

	// Чтение файла в память
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	// Преобразование содержимого файла в строки
	lines := strings.Split(string(fileContent), "\n")

	// Создание нового списка строк, в который будут записаны все строки, кроме той, что содержит нужный ID
	var updatedLines []string
	found := false

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("(%d,", userID)) {
			// Логирование и пропуск строки с нужным ID
			logger.Info("Removing migration line: %s", line)
			found = true
			continue
		}
		updatedLines = append(updatedLines, line)
	}

	if !found {
		logger.Warning("User with ID %d not found in migration file.", userID)
		return nil
	}

	// Запись обновленных данных обратно в файл
	output := strings.Join(updatedLines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("could not write updated SQL file: %v", err)
	}

	logger.Info("Migration file updated successfully")
	return nil
}

func removeTaskFromMigrationFile(userID int) error {
	filePath := "migrations/20230707120000_insert_initial_tasks.sql"

	// Чтение файла в память
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	// Преобразование содержимого файла в строки
	lines := strings.Split(string(fileContent), "\n")

	// Создание нового списка строк без строк, которые нужно удалить
	var updatedLines []string
	for _, line := range lines {
		// Проверка и удаление строк, содержащих вставки задач с указанным userID
		if strings.Contains(line, "INSERT INTO tasks") || strings.Contains(line, fmt.Sprintf("(%d,", userID)) {
			logger.Info("Removing line: %s", line)
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	// Запись обновленных данных обратно в файл
	output := strings.Join(updatedLines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("could not write SQL file: %v", err)
	}

	logger.Info("Migration file updated successfully")
	return nil
}
