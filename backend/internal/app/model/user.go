package model

import (
	"time"
)

type User struct {
	ID           uint          `gorm:"primaryKey" json:"user_id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Exist        bool          `json:"exist"`
	Password     string        `json:"password"`
	Status       string        `json:"status"`
	DoFlag       time.Time     `json:"do_flag"`
	FlagNumber   int           `json:"flag_number"`
	Count        int           `json:"count"`
	DaKaNumber   []Daka_number `grom:"foreignKey" `
	LearnTimes   []LearnTime   `gorm:"foreignKey:UserID"`  //外键绑定learn_time表
	Flags        []Flag        `gorm:"foreignKey:UserID"`  //外键绑定flag表
	Posts        []Post        `gorm:"foreignKey:UserID"`  //外键绑定post表
	Achievements []Achievement `gorm:"foreignKey:UserID;"` //多对多绑定achievement表
}

// Flag
type Flag struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Flag           string    `json:"flag"`
	PlanContent    string    `json:"plan_content"`
	Label          string    `json:"label"`
	Priority       int       `json:"priority"`
	UserID         uint      `json:"user_id"`
	IsHiden        bool      `json:"is_hiden"`
	HadDone        bool      `json:"had_done"`             //是否完成
	DoneNumber     int       `json:"done_number"`          //已完成程度
	PlanDoneNumber int       `json:"plan_done_number"`     //目标程度
	AgreeNumber    int       `json:"agree_number"`         //点赞数量
	Comments       []Comment `gorm:"foreignKey:CommentID"` //外键绑定comment表
	CreatedAt      time.Time `json:"created_at"`           //创建时间
	StartTime      time.Time `json:"start_time"`           //开始时间
	DeadTime       time.Time `json:"time"`                 //结束时间
}

// 帖子
type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    uint      `gorm:"fori" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `gorm:"foreignKey:CommentID"` //外键绑定post_comment表
}

// 帖子评论
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CommentID uint      `json:"comment_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Achievement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	UserID      uint      `json:"user_id"`
	Description string    `json:"description"`
	HadDone     bool      `json:"had_done"`
	GotTime     time.Time `json:"got_time"`
}

type LearnTime struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `json:"user_id"`
	Duration  int       `json:"duration"` // 学习时长，单位为分钟
}

type Daka_number struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	HadDone   bool      `json:"had_done"`
	MonthDaka int       `json:"month_daka"`
	DaKaDate  time.Time `json:"daka_date"`
}

// 邮箱验证码
type EmailCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email"`
	HadUse    bool      `json:"had_use"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	Expires   time.Time `json:"expires"`
}
