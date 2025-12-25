package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
)

// 一言API响应结构
type HitokotoResponse struct {
	ID         int    `json:"id"`
	Hitokoto   string `json:"hitokoto"` // 一言正文
	Type       string `json:"type"`     // 类型
	From       string `json:"from"`     // 来源
	FromWho    string `json:"from_who"` // 作者
	Creator    string `json:"creator"`  // 添加者
	CreatorUID int    `json:"creator_uid"`
	Reviewer   int    `json:"reviewer"`
	UUID       string `json:"uuid"`
	CommitFrom string `json:"commit_from"`
	CreatedAt  string `json:"created_at"`
	Length     int    `json:"length"`
}

const (
	HitokotoAPIURL     = "https://v1.hitokoto.cn"
	HitokotoUserID     = 9999 // 一言系统账户ID
	HitokotoUsername   = "一言"
	HitokotoUserAvatar = 3 // 头像索引为3（avatar3）
)

// 获取一言内容
func fetchHitokoto() (*HitokotoResponse, error) {
	resp, err := http.Get(HitokotoAPIURL)
	if err != nil {
		return nil, fmt.Errorf("请求一言API失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var hitokoto HitokotoResponse
	if err := json.Unmarshal(body, &hitokoto); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &hitokoto, nil
}

// 确保一言系统账户存在
func ensureHitokotoUser() error {
	// 检查数据库连接
	if repository.DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 检查是否已存在
	var user model.User
	result := repository.DB.Where("id = ?", HitokotoUserID).First(&user)

	if result.Error == nil {
		// 用户已存在
		return nil
	}

	// 创建一言系统账户
	hitokotoUser := model.User{
		ID:       HitokotoUserID,
		Name:     HitokotoUsername,
		Email:    "hitokoto@system.local",
		Password: "", // 系统账户无需密码
		Status:   "一言·每日智慧",
		HeadShow: HitokotoUserAvatar,
		IsRemind: false,
	}

	if err := repository.DB.Create(&hitokotoUser).Error; err != nil {
		return fmt.Errorf("创建一言账户失败: %v", err)
	}

	return nil
}

// 发布一言帖子
func postHitokoto() error {
	// 获取一言内容
	hitokoto, err := fetchHitokoto()
	if err != nil {
		return err
	}

	// 构建帖子内容
	var content string
	if hitokoto.FromWho != "" && hitokoto.From != "" {
		content = fmt.Sprintf("%s\n\n—— %s《%s》", hitokoto.Hitokoto, hitokoto.FromWho, hitokoto.From)
	} else if hitokoto.From != "" {
		content = fmt.Sprintf("%s\n\n—— 《%s》", hitokoto.Hitokoto, hitokoto.From)
	} else if hitokoto.FromWho != "" {
		content = fmt.Sprintf("%s\n\n—— %s", hitokoto.Hitokoto, hitokoto.FromWho)
	} else {
		content = hitokoto.Hitokoto
	}

	// 创建帖子
	post := model.Post{
		Title:   "每日一言",
		Content: content,
		UserID:  HitokotoUserID,
	}

	if err := repository.DB.Create(&post).Error; err != nil {
		return fmt.Errorf("发布一言帖子失败: %v", err)
	}

	return nil
}

// 启动一言定时任务
func StartHitokotoScheduler() {
	// 确保一言系统账户存在
	if err := ensureHitokotoUser(); err != nil {
		return
	}

	// 启动定时器
	go func() {
		for {
			now := time.Now()
			// 计算下次8点的时间
			next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
			if now.After(next) {
				// 如果现在已经过了今天的8点，则设置为明天8点
				next = next.Add(24 * time.Hour)
			}

			// 等待到8点
			duration := next.Sub(now)

			time.Sleep(duration)

			// 发布一言
			postHitokoto()

			// 等待1分钟，避免重复执行
			time.Sleep(1 * time.Minute)
		}
	}()
}
