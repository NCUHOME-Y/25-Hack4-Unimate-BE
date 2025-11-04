package model

import (
	"time"
)

type User struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Email    string
	Password string
}

type Flag struct {
	ID          uint `gorm:"primaryKey"`
	Flag        string
	PlanContent string
	UserID      uint
	IsHiden     bool
	time        time.Time
}
