package models

import "time"

type User struct {
	ID             int       `json:"id"`
	PassportNumber string    `json:"passportNumber"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Surname        string    `json:"surname"`
	Name           string    `json:"name"`
	Patronymic     string    `json:"patronymic"`
	Address        string    `json:"address"`
}

type Task struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Name      string    `json:"name"`
	Hours     int       `json:"hours"`
	Minutes   int       `json:"minutes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
