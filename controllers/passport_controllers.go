package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"test-project/logger"
)

func ValidatePassportNumber(passportNumber string, w http.ResponseWriter) (string, string, error) {
	passportParts := strings.Split(passportNumber, " ")
	if len(passportParts) != 2 {
		logger.Warning("Invalid passport format")
		http.Error(w, "Invalid passport format", http.StatusBadRequest)
		return "", "", fmt.Errorf("invalid passport format")
	}
	passportSerieStr := passportParts[0]
	passportNumberStr := passportParts[1]
	logger.Info("Parsed passport number: series=%s, number=%s", passportSerieStr, passportNumberStr)

	_, err := strconv.Atoi(passportSerieStr)
	if err != nil || len(passportSerieStr) != 4 {
		logger.Error("Invalid passport series format: %v", passportSerieStr)
		http.Error(w, "Invalid passport series format", http.StatusBadRequest)
		return "", "", fmt.Errorf("invalid passport series format")
	}

	_, err = strconv.Atoi(passportNumberStr)
	if err != nil || len(passportNumberStr) != 6 {
		logger.Error("Invalid passport number format: %v", passportNumberStr)
		http.Error(w, "Invalid passport number format", http.StatusBadRequest)
		return "", "", fmt.Errorf("invalid passport number format")
	}

	return passportSerieStr, passportNumberStr, nil
}
