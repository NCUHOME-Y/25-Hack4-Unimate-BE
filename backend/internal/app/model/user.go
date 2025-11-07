package model

import (
	"time"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Status   string `json:"status"`
	Flags    []Flag `gorm:"foreignKey:UserID"`   //外键绑定flag表
	Posts    []Post `gorm:"foreignKey:AuthorID"` //外键绑定post表
}

type Flag struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Flag           string    `json:"flag"`
	PlanContent    string    `json:"plan_content"`
	UserID         uint      `json:"user_id"`
	IsHiden        bool      `json:"is_hiden"`
	HadDone        bool      `json:"had_done"`         //是否完成
	DoneNumber     int       `json:"done_number"`      //已完成程度
	PlanDoneNumber int       `json:"plan_done_number"` //目标程度
	DeadTime       time.Time `json:"time"`             //结束时间
}

type Post struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	AuthorID  uint          `gorm:"fori" json:"author_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Comments  []PostComment `gorm:"foreignKey:PostID"` //外键绑定post_comment表
}

type PostComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `json:"post_id"`
	Content   string    `json:"content"`
	AuthorID  uint      `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
