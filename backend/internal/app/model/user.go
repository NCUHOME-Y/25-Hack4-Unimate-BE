package model

import (
	"time"
)

type User struct {
	ID             uint          `gorm:"primaryKey" json:"user_id"`       //用户ID
	Name           string        `json:"name"`                            //用户名
	Email          string        `json:"email"`                           //邮箱
	Exist          bool          `json:"exist"`                           //邮箱是否验证
	Password       string        `json:"password"`                        //密码
	Status         string        `json:"status"`                          //用户状态
	IsRemind       bool          `json:"is_remind" gorm:"default:true"`   //是否开启提醒
	DoFlag         time.Time     `json:"do_flag"`                         //最后打卡时间
	HeadShow       int           `json:"head_show" gorm:"default:1"`      //头像显示
	RemindHour     int           `json:"time_remind" default:"12"`        //提醒小时
	RemindMin      int           `json:"min_remind" default:"0"`          //提醒分钟
	Daka           int           `json:"daka"`                            //总打卡数
	MonthLearntime int           `json:"month_learn_time"`                //本月学习时长
	FlagNumber     int           `json:"flag_number"`                     //完成flag数量
	Count          int           `json:"count"`                           //积分
	Labels         Label         `json:"labels" gorm:"foreignKey:UserID"` //完成flag的标签数
	DaKaNumber     []Daka_number `grom:"foreignKey" `
	LearnTimes     []LearnTime   `gorm:"foreignKey:UserID"`  //外键绑定learn_time表
	Flags          []Flag        `gorm:"foreignKey:UserID"`  //外键绑定flag表
	Posts          []Post        `gorm:"foreignKey:UserID"`  //外键绑定post表
	Achievements   []Achievement `gorm:"foreignKey:UserID;"` //多对多绑定achievement表
}

// Flag
type Flag struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	Flag           string        `json:"flag"`
	PlanContent    string        `json:"plan_content"`
	Label          string        `json:"label"`
	Priority       int           `json:"priority"`
	UserID         uint          `json:"user_id"`
	IsHiden        bool          `json:"is_hiden"`
	HadDone        bool          `json:"had_done"`          //是否完成
	DoneNumber     int           `json:"done_number"`       //已完成程度
	PlanDoneNumber int           `json:"plan_done_number"`  //目标程度
	Like           int           `json:"like"`              //点赞数量
	FlagComments   []FlagComment `gorm:"foreignKey:FlagID"` //外键绑定comment表
	CreatedAt      time.Time     `json:"created_at"`        //创建时间
	StartTime      time.Time     `json:"start_time"`        //开始时间
	DeadTime       time.Time     `json:"time"`              //结束时间
}

// 帖子
type Post struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Like      int           `json:"like"`
	UserID    uint          `gorm:"fori" json:"user_id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Comments  []PostComment `gorm:"foreignKey:PostID"` //外键绑定post_comment表
}

// 帖子评论
type PostComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `json:"post_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// flag评论
type FlagComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FlagID    uint      `json:"flag_id"`
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

// 标签
type Label struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `json:"user_id"`
	Life   int  `json:"life"`
	Study  int  `json:"study"`
	Work   int  `json:"work"`
	Like   int  `json:"like"`
	Sport  int  `json:"sport"`
}
