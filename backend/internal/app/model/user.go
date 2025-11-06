package model

import (
	"time"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Flags    []Flag `gorm:"foreignKey:UserID"`
}

type Flag struct {
	ID             uint      `gorm:"primaryKey"`
	Flag           string    `json:"flag"`
	PlanContent    string    `json:"plan_content"`
	UserID         uint      `json:"user_id"`
	IsHiden        bool      `json:"is_hiden"`
	HadDone        bool      `json:"had_done"`
	DoneNumber     int       `json:"done_number"`
	PlanDoneNumber int       `json:"plan_done_number"`
	DeadTime       time.Time `json:"time"`
}
