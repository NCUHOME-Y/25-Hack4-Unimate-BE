package model

import (
	"time"

	"gorm.io/gorm"
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

// Flag - 前端字段为主
type Flag struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Title     string        `gorm:"column:flag" json:"title"`          // 前端: title
	Detail    string        `gorm:"column:plan_content" json:"detail"` // 前端: detail
	LabelStr  string        `gorm:"column:label" json:"-"`             // 数据库字段（字符串）
	Label     int           `gorm:"-" json:"label"`                    // 前端: label (1-5数字)
	Priority  int           `json:"priority"`                          // 前端: priority (1-4)
	UserID    uint          `json:"user_id"`
	IsHidden  bool          `gorm:"column:is_hiden;not null;default:false" json:"-"` // 数据库字段（不导出到JSON）
	IsPublic  bool          `gorm:"-" json:"is_public"`                              // 前端字段（不存储到数据库，通过 AfterFind 计算）
	Completed bool          `gorm:"column:had_done" json:"completed"`                // 前端: completed
	Count     int           `gorm:"column:done_number" json:"count"`                 // 前端: count (已完成次数)
	Total     int           `gorm:"column:plan_done_number" json:"total"`            // 前端: total (目标次数)
	Points    int           `json:"points"`                                          // 前端: points (积分)
	Likes     int           `gorm:"column:like" json:"likes"`                        // 前端: agreeNumber → likes
	Comments  []FlagComment `gorm:"foreignKey:FlagID" json:"comments"`               // 评论列表
	CreatedAt time.Time     `json:"created_at"`                                      // 前端: createdAt
	StartTime time.Time     `json:"start_time"`                                      // 前端: startTime
	EndTime   time.Time     `gorm:"column:time" json:"end_time"`                     // 前端: endTime
}

// AfterFind - GORM钩子：查询后自动将 IsHidden 反转为 IsPublic，并转换label
func (f *Flag) AfterFind(tx *gorm.DB) error {
	f.IsPublic = !f.IsHidden

	// 将字符串label转换为数字（统一前后端格式）
	labelMap := map[string]int{
		"life":  1,
		"study": 2,
		"work":  3,
		"like":  4,
		"sport": 5,
		"生活":    1,
		"学习":    2,
		"工作":    3,
		"兴趣":    4,
		"运动":    5,
	}
	if val, ok := labelMap[f.LabelStr]; ok {
		f.Label = val
	} else {
		f.Label = 2 // 默认学习
	}

	return nil
}

// BeforeSave - GORM钩子：保存前将 IsPublic 反转为 IsHidden，并转换label
func (f *Flag) BeforeSave(tx *gorm.DB) error {
	f.IsHidden = !f.IsPublic

	// 将数字label转换为字符串存储到数据库
	labelMap := map[int]string{
		1: "生活",
		2: "学习",
		3: "工作",
		4: "兴趣",
		5: "运动",
	}
	if val, ok := labelMap[f.Label]; ok {
		f.LabelStr = val
	} else {
		f.LabelStr = "学习" // 默认学习
	}

	return nil
}

// 帖子
type Post struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Like       int           `json:"like"`
	UserID     uint          `gorm:"foreignKey:UserID" json:"user_id"`
	User       *User         `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联用户信息
	UserName   string        `gorm:"-" json:"userName"`                       // 前端需要的用户名（计算字段）
	UserAvatar string        `gorm:"-" json:"userAvatar"`                     // 前端需要的用户头像（计算字段）
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Comments   []PostComment `gorm:"foreignKey:PostID" json:"comments"` //外键绑定post_comment表
}

// AfterFind - GORM钩子：查询后自动填充用户信息
func (p *Post) AfterFind(tx *gorm.DB) error {
	if p.User != nil {
		p.UserName = p.User.Name
		if p.User.HeadShow > 0 && p.User.HeadShow <= 6 {
			// 映射到前端 assets 的图片路径
			avatarFiles := []string{"131601", "131629", "131937", "131951", "132014", "133459"}
			p.UserAvatar = "/src/assets/images/screenshot_20251114_" + avatarFiles[p.User.HeadShow-1] + ".png"
		} else {
			p.UserAvatar = ""
		}
	}
	return nil
}

// 帖子评论
type PostComment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PostID     uint      `json:"post_id"`
	UserID     uint      `json:"userId" gorm:"column:user_id"` // 评论者ID
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	User       *User     `gorm:"foreignKey:UserID" json:"-"` // 关联用户信息
	UserName   string    `gorm:"-" json:"userName"`          // 前端需要的用户名（计算字段）
	UserAvatar string    `gorm:"-" json:"userAvatar"`        // 前端需要的用户头像（计算字段）
}

// AfterFind - GORM钩子：查询后自动填充用户信息
func (c *PostComment) AfterFind(tx *gorm.DB) error {
	if c.User != nil {
		c.UserName = c.User.Name
		if c.User.HeadShow > 0 && c.User.HeadShow <= 6 {
			avatarFiles := []string{"131601", "131629", "131937", "131951", "132014", "133459"}
			c.UserAvatar = "/src/assets/images/screenshot_20251114_" + avatarFiles[c.User.HeadShow-1] + ".png"
		} else {
			c.UserAvatar = ""
		}
	}
	return nil
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
	UserID    uint      `gorm:"column:user_id" json:"user_id"`
	HadDone   bool      `gorm:"column:had_done" json:"had_done"`
	MonthDaka int       `gorm:"column:month_daka" json:"month_daka"`
	DaKaDate  time.Time `gorm:"column:daka_date" json:"daka_date"`
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

// 埋点
type TrackPoint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
}

// 聊天消息
type ChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `json:"from" gorm:"column:from_user_id"`
	ToUserID   uint      `json:"to" gorm:"column:to_user_id"` // 0表示群聊
	RoomID     string    `json:"room_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	User       *User     `gorm:"foreignKey:FromUserID" json:"-"` // 关联发送者信息
	UserName   string    `gorm:"-" json:"user_name"`
	UserAvatar string    `gorm:"-" json:"user_avatar"`
}

// AfterFind - GORM钩子：查询后自动填充用户信息
func (m *ChatMessage) AfterFind(tx *gorm.DB) error {
	if m.User != nil {
		m.UserName = m.User.Name
		if m.User.HeadShow > 0 && m.User.HeadShow <= 6 {
			avatarFiles := []string{"131601", "131629", "131937", "131951", "132014", "133459"}
			m.UserAvatar = "/src/assets/images/screenshot_20251114_" + avatarFiles[m.User.HeadShow-1] + ".png"
		}
	}
	return nil
}
