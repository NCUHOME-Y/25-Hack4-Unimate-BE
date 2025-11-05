package model

import (
	"time"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json :name`
	Email    string `json:email`
	Password string `json:password`
}

type Flag struct {
	ID          uint      `gorm:"primaryKey"`
	Flag        string    `json:"flag"`
	PlanContent string    `json:"plan_content"`
	UserID      uint      `json:"user_id"`
	IsHiden     bool      `json:"is_hiden"`
	time        time.Time `json:"time"`
}
