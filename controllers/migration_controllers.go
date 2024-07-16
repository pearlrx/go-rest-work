package controllers

import (
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"test-project/logger"
	"test-project/models"
	"time"
)

var (
	filePathUserMigration = "migrations/20230707120000_insert_initial_users.sql"
	filePathTaskMigration = "migrations/20230707120000_insert_initial_tasks.sql"
)

func addTaskToMigrationFile(task models.Task) {
	filePath := filePathTaskMigration
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
	filePath := filePathUserMigration
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

	if _, err = file.WriteString(migrationLine); err != nil {
		logger.Error("Failed to write to migration file: %v", err)
	}
}

func updateUserInMigrationFile(updatedUser models.User, id int) {
	filePath := filePathUserMigration

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Warning("Failed to read migration file: %v", err)
		return
	}

	lines := strings.Split(string(fileContent), "\n")

	newMigrationLine := fmt.Sprintf(
		"INSERT INTO users (id, passport_number, surname, name, patronymic, address, created_at, updated_at) VALUES (%d, '%s', '%s', '%s', '%s', '%s', NOW(), NOW());",
		id, updatedUser.PassportNumber, updatedUser.Surname, updatedUser.Name, updatedUser.Patronymic, updatedUser.Address,
	)

	updated := false
	for i, line := range lines {
		if strings.Contains(line, fmt.Sprintf("(%d,", id)) {
			lines[i] = newMigrationLine
			updated = true
			break
		}
	}

	if !updated {
		logger.Warning("User with ID %d not found in migration file.", id)
		return
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		logger.Error("Failed to write to migration file: %v", err)
	} else {
		logger.Info("Successfully updated migration line for user %s %s", updatedUser.Name, updatedUser.Surname)
	}
}

func removeUserFromMigrationFile(userID int) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	filePath := filePathUserMigration

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	lines := strings.Split(string(fileContent), "\n")

	var updatedLines []string
	found := false

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("(%d,", userID)) {
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

	output := strings.Join(updatedLines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("could not write updated SQL file: %v", err)
	}

	logger.Info("Migration file updated successfully")
	return nil
}

func removeTaskFromMigrationFile(userID int) error {
	filePath := filePathTaskMigration

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read SQL file: %v", err)
	}

	lines := strings.Split(string(fileContent), "\n")

	var updatedLines []string
	for _, line := range lines {
		if strings.Contains(line, "INSERT INTO tasks") || strings.Contains(line, fmt.Sprintf("(%d,", userID)) {
			logger.Info("Removing line: %s", line)
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	output := strings.Join(updatedLines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("could not write SQL file: %v", err)
	}

	logger.Info("Migration file updated successfully")
	return nil
}
