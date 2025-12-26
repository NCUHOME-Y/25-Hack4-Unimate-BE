package model

import (
	"time"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"gorm.io/gorm"
)

type User struct {
	ID         uint       `gorm:"primaryKey" json:"userId"`      //用户ID
	Name       string     `json:"name"`                          //用户名
	Email      string     `json:"email"`                         //邮箱
	Password   string     `json:"password"`                      //密码
	Status     string     `json:"status"`                        //用户状态
	IsRemind   bool       `json:"isRemind" gorm:"default:false"` //是否开启提醒
	DoFlag     *time.Time `json:"doFlag"`                        //最后打卡时间
	HeadShow   int        `json:"headShow" gorm:"default:1"`     //头像显示
	RemindHour int        `json:"timeRemind" default:"12"`       //兼容字段：提醒小时（旧）
	RemindMin  int        `json:"minRemind" default:"0"`         //兼容字段：提醒分钟（旧）

	// 新增：将学习提醒和 Flag 提醒独立管理
	IsStudyRemind   bool          `json:"isStudyRemind" gorm:"default:false"` // 是否开启学习提醒
	StudyRemindHour int           `json:"studyRemindHour" gorm:"default:12"`  // 学习提醒小时
	StudyRemindMin  int           `json:"studyRemindMin" gorm:"default:0"`    // 学习提醒分钟
	IsFlagRemind    bool          `json:"isFlagRemind" gorm:"default:false"`  // 是否开启 Flag 提醒（用户级总开关）
	Daka            int           `json:"daka"`                               //总打卡数
	MonthLearntime  int           `json:"monthLearnTime"`                     //本月学习时长
	FlagNumber      int           `json:"flagNumber"`                         //完成flag数量
	Count           int           `json:"count"`                              //积分
	Labels          Label         `json:"labels" gorm:"foreignKey:UserID"`    //完成flag的标签数
	DaKaNumber      []Daka_number `gorm:"foreignKey:UserID"`
	LearnTimes      []LearnTime   `gorm:"foreignKey:UserID"` //外键绑定learn_time表
	Flags           []Flag        `gorm:"foreignKey:UserID"` //外键绑定flag表
	Posts           []Post        `gorm:"foreignKey:UserID"` //外键绑定post表
	Achievements    []Achievement `gorm:"foreignKey:UserID"` //一对多绑定achievement表
	AIHistories     []AIHistory   `gorm:"foreignKey:UserID"` //一对多绑定ai_history表
}

// Flag - 前端字段为主
type Flag struct {
	ID                 uint          `gorm:"primaryKey" json:"id"`
	Title              string        `gorm:"column:flag" json:"title"`          // 前端: title
	Detail             string        `gorm:"column:plan_content" json:"detail"` // 前端: detail
	Label              int           `gorm:"column:label" json:"label"`         // 前端&数据库: label (1-5数字)
	Priority           int           `json:"priority"`                          // 前端: priority (1-4)
	UserID             uint          `json:"userId"`
	PostID             *uint         `gorm:"column:post_id;index" json:"postId,omitempty"`                       // 关联的社交帖子ID（null=未分享，有值=已分享且指向Post.ID）
	Completed          bool          `gorm:"column:had_done" json:"completed"`                                   // 前端: completed
	Count              int           `gorm:"column:done_number" json:"count"`                                    // 前端: count (已完成次数)
	DailyTotal         int           `gorm:"column:daily_total" json:"total"`                                    // 前端: total (每日所需完成次数)
	Points             int           `json:"points"`                                                             // 前端: points (积分)
	Likes              int           `gorm:"column:like" json:"likes"`                                           // 前端: agreeNumber → likes
	Comments           []FlagComment `gorm:"foreignKey:FlagID" json:"comments"`                                  // 评论列表
	CreatedAt          time.Time     `json:"createdAt"`                                                          // 前端: createdAt
	StartTime          *time.Time    `gorm:"column:start_time" json:"startTime"`                                 // 前端: startTime
	EndTime            *time.Time    `gorm:"column:end_time" json:"endTime"`                                     // 前端: endTime
	EnableNotification bool          `gorm:"column:enable_notification;default:false" json:"enableNotification"` // 是否启用该flag的消息提醒
	ReminderTime       string        `gorm:"column:reminder_time;default:'12:00'" json:"reminderTime"`           // 该flag的提醒时间 (HH:MM 格式)
}

// 注意：Label字段已改为直接存储数字(1-5)到数据库，不再需要类型转换钩子

// 帖子
type Post struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Like       int           `json:"likes"`                                   // 前端期望 likes
	UserID     uint          `gorm:"foreignKey:UserID" json:"userId"`         // 驼峰
	FlagID     *uint         `gorm:"index" json:"flagId,omitempty"`           // 驼峰
	User       *User         `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联用户信息
	UserName   string        `gorm:"-" json:"userName"`                       // 前端需要的用户名（计算字段）
	UserAvatar string        `gorm:"-" json:"userAvatar"`                     // 前端需要的用户头像（计算字段）
	CreatedAt  time.Time     `json:"createdAt"`                               // 驼峰
	UpdatedAt  time.Time     `json:"updatedAt"`                               // 驼峰
	Comments   []PostComment `gorm:"foreignKey:PostID" json:"comments"`       //外键绑定post_comment表
}

// AfterFind - GORM钩子：查询后自动填充用户信息
func (p *Post) AfterFind(tx *gorm.DB) error {
	if p.User != nil {
		p.UserName = p.User.Name
		p.UserAvatar = utils.GetAvatarPath(p.User.HeadShow)
	}
	return nil
}

// 帖子评论
type PostComment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PostID     uint      `json:"postId"`                       // 驼峰
	UserID     uint      `json:"userId" gorm:"column:user_id"` // 评论者ID，驼峰
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`                  // 驼峰
	UpdatedAt  time.Time `json:"updatedAt"`                  // 驼峰
	User       *User     `gorm:"foreignKey:UserID" json:"-"` // 关联用户信息
	UserName   string    `gorm:"-" json:"userName"`          // 前端需要的用户名（计算字段）
	UserAvatar string    `gorm:"-" json:"userAvatar"`        // 前端需要的用户头像（计算字段）
}

// AfterFind - GORM钩子：查询后自动填充用户信息
func (c *PostComment) AfterFind(tx *gorm.DB) error {
	if c.User != nil {
		c.UserName = c.User.Name
		c.UserAvatar = utils.GetAvatarPath(c.User.HeadShow)
	}
	return nil
}

// 用户-帖子点赞关系表
type UserPostLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	PostID    uint      `gorm:"index" json:"postId"`
	CreatedAt time.Time `json:"createdAt"`
}

// flag评论
type FlagComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FlagID    uint      `json:"flagId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Achievement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	UserID      uint      `gorm:"index;not null" json:"userId"`
	Description string    `json:"description"`
	HadDone     bool      `json:"hadDone"`
	GotTime     time.Time `json:"gotTime"`
	User        *User     `gorm:"foreignKey:UserID;references:ID" json:"-"` // 补充关联声明
}

type LearnTime struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    uint      `json:"userId"`
	Duration  int       `json:"duration"` // 学习时长，单位为秒 (Duration in seconds)
}

type Daka_number struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"column:user_id" json:"userId"`
	HadDone   bool      `gorm:"column:had_done" json:"hadDone"`
	MonthDaka int       `gorm:"column:month_daka" json:"monthDaka"`
	DaKaDate  time.Time `gorm:"column:daka_date" json:"dakaDate"`
}

// 邮箱验证码
type EmailCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email"`
	HadUse    bool      `json:"hadUse"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
	Expires   time.Time `json:"expires"`
}

// 标签
type Label struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `json:"userId"`
	Life   int  `json:"life"`
	Study  int  `json:"study"`
	Work   int  `json:"work"`
	Like   int  `json:"likeCount"`
	Sport  int  `json:"sport"`
}

// 埋点
type TrackPoint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"userId"`
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
}

// 聊天消息
type ChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `json:"from" gorm:"column:from_user_id"`
	ToUserID   uint      `json:"to" gorm:"column:to_user_id"` // 0表示群聊
	RoomID     string    `json:"roomId"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	User       *User     `gorm:"foreignKey:FromUserID" json:"-"` // 关联发送者信息
	UserName   string    `gorm:"-" json:"userName"`
	UserAvatar string    `gorm:"-" json:"userAvatar"`
}

// AfterFind - GORM钩子：查询后自动填充用户信息
func (m *ChatMessage) AfterFind(tx *gorm.DB) error {
	if m.User != nil {

		m.UserName = m.User.Name
		m.UserAvatar = utils.GetAvatarPath(m.User.HeadShow)
	}
	return nil
}

// 积分日志（记录每次积分变动，用于统计“今日获得积分”）
type PointsLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

// AI历史记录
type AIHistory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"userId"`
	Background    string    `json:"background"`    // 个人背景
	Goal          string    `json:"goal"`          // 目标
	Difficulty    string    `json:"difficulty"`    // 难度
	GeneratedPlan string    `json:"generatedPlan"` // 生成的计划（JSON格式）
	CreatedAt     time.Time `json:"createdAt"`
}
