package repository

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/sirupsen/logrus"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

// 链接数据库
func DBconnect() error {
	dsn := os.Getenv("DB_DSN")

	// 如果没有 DB_DSN，则从分散的环境变量构建
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "3306"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "root"
		}
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "unimate"
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, port, dbname)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.LogError("数据库连接失败", logrus.Fields{"error": err})
		return fmt.Errorf("数据库连接失败: %v", err)
	}
	DB = db

	// 自动迁移数据表
	if err := DB.AutoMigrate(&model.User{}, &model.Flag{}, &model.Post{}, &model.PostComment{}, &model.Achievement{}, &model.LearnTime{}, &model.Daka_number{}, &model.EmailCode{}, &model.FlagComment{}, &model.TrackPoint{}, &model.ChatMessage{}, &model.UserPostLike{}, &model.PointsLog{}, &model.AIHistory{}); err != nil {
		utils.LogError("数据库表迁移失败", logrus.Fields{"error": err})
		return fmt.Errorf("数据库表迁移失败: %v", err)
	}

	utils.LogInfo("数据库连接成功", nil)
	return nil
}

// user添加到数据库
func AddUserToDB(user *model.User) error {
	result := DB.Create(user)
	return result.Error
}

// flag添加到数据库
func AddFlagToDB(Id uint, flag model.Flag) error {
	flag.UserID = Id
	result := DB.Create(&flag)
	return result.Error
}

// 更新flag的完整信息
func UpdateFlag(flagID uint, updates map[string]interface{}) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Updates(updates)
	return result.Error
}

// 通过邮箱删除用户
func DeleteUserByEmail(email string) error {
	result := DB.Where("email = ?", email).Delete(&model.User{})
	return result.Error
}

// 更新用户信息
func UpdateUser(user model.User) error {
	result := DB.Save(&user)
	return result.Error
}

// 从数据库删除flag
func DeleteFlagFromDB(flagID uint) error {
	result := DB.Delete(&model.Flag{}, flagID)
	return result.Error
}

// 通过用户ID获取flag列表
func GetFlagsByUserID(userID uint) ([]model.Flag, error) {
	var flags []model.Flag
	// 只返回当天可用的flag: 每天 或 在起止日期范围内
	today := time.Now()
	result := DB.Where("user_id = ?", userID).
		Where("(start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", today, today).
		Order("priority").
		Find(&flags)
	return flags, result.Error
}

// 获取有起始日期且未过期的flag（用于日历高亮）
func GetFlagsWithDatesByUserID(userID uint, today time.Time) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ? AND start_time IS NOT NULL AND (end_time IS NULL OR end_time >= ?)", userID, today).Find(&flags)
	return flags, result.Error
}

// 获取预设flag（未到起始日期且未过期）
func GetPresetFlagsByUserID(userID uint, today time.Time) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ? AND start_time IS NOT NULL AND start_time > ? AND (end_time IS NULL OR end_time >= ?)", userID, today, today).
		Order("start_time").
		Find(&flags)
	return flags, result.Error
}

// 获取过期flag
func GetExpiredFlagsByUserID(userID uint, today time.Time) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ? AND end_time < ?", userID, today).
		Order("end_time desc").
		Limit(6).
		Find(&flags)
	return flags, result.Error
}

// 通过用户邮箱获取用户（完整版）
func GetUserByEmail(Email string) (model.User, error) {
	var user model.User
	result := DB.Preload("Achievements").Preload("Flags").Preload("Posts").Where("email = ?", Email).First(&user)
	return user, result.Error
}

// 🔧 性能优化：轻量级邮箱查询 - 用于登录验证
func GetUserBasicByEmail(Email string) (model.User, error) {
	var user model.User
	result := DB.Where("email = ?", Email).First(&user)
	return user, result.Error
}

// 通过用户名获取用户
func GetUserByName(name string) (model.User, error) {
	var user model.User
	result := DB.Where("name = ?", name).First(&user)
	return user, result.Error
}

// 通过用户ID获取用户（完整版 - 包含所有关联数据）
func GetUserByID(userID uint) (model.User, error) {
	var user model.User
	result := DB.Preload("Achievements").Preload("Flags").Preload("Posts").Where("id = ?", userID).First(&user)
	return user, result.Error
}

// 🔧 性能优化：轻量级用户查询 - 不加载关联数据
func GetUserBasicByID(userID uint) (model.User, error) {
	var user model.User
	result := DB.Where("id = ?", userID).First(&user)
	return user, result.Error
}

// 搜索关键词查询用户，可以是邮箱是用户名
func SearchUsers(keyword string) ([]model.User, error) {
	var users []model.User
	like := "%" + keyword + "%"
	err := DB.Preload("Achievements").
		Preload("Flags").
		Preload("Posts").
		Where("name LIKE ? OR email LIKE ?", like, like).Find(&users).Error // 把 Flags 一起查出来
	return users, err
}

// 更新flag的可见性
func UpdateFlagVisibility(flagID uint, isHidden bool) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("is_public", !isHidden)
	return result.Error
}

// 更新flag的内容
func UpdateFlagContent(flagID uint, newContent string) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("flag", newContent)
	return result.Error
}

// 更新flag的计划内容
func UpdatePlanContent(flagID uint, newPlanContent string) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("plan_content", newPlanContent)
	return result.Error
}

// 更新flag的评论
func UpdateFlagComment(flagID uint, newComment string) error {
	var flagComment model.FlagComment
	flagComment.FlagID = flagID
	flagComment.Content = newComment
	result := DB.Model(&model.FlagComment{}).Where("flag_id = ?", flagID).Create(&flagComment)
	return result.Error
}

// 删除flag的评论
func DeleteFlagComment(flagcommentID uint) error {
	result := DB.Model(&model.FlagComment{}).Where("id = ?", flagcommentID).Delete(&model.FlagComment{})
	return result.Error
}

// 更新用户密码
func UpdatePassword(id uint, newPassword string) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("Password", newPassword)
	return result.Error
}

// 通过邮箱更新密码
func UpdatePasswordByEmail(email string, newPassword string) error {
	result := DB.Model(&model.User{}).Where("email=?", email).Update("Password", newPassword)
	return result.Error
}

// 更新用户名
func UpdateUserName(id uint, newName string) error {
	// 使用 map 更新确保列名和大小写问题不会导致 SQL 错误
	result := DB.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{"name": newName})
	return result.Error
}

// 更新flag的完成数量
func UpdateFlagDoneNumber(flagID uint, doneNumber int) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("done_number", doneNumber)
	return result.Error
}

// 更新flag的完成状态
func UpdateFlagHadDone(flagID uint, isdo bool) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("had_done", isdo)
	return result.Error
}

// 更新flag的提醒状态
func UpdateFlagNotification(flagID uint, enabled bool) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("enable_notification", enabled)
	return result.Error
}

// 统计用户已启用提醒的flag数量（只统计未完成的flag）
func CountEnabledNotificationFlags(userID uint) (int64, error) {
	var count int64
	result := DB.Model(&model.Flag{}).Where("user_id = ? AND enable_notification = ? AND had_done = ?", userID, true, false).Count(&count)
	return count, result.Error
}

// 打卡时间更新
func UpdateUserDoFlag(id uint, doFlag time.Time) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("do_flag", doFlag)
	return result.Error
}

// 更新flag的完成期限
func UpdateFlagDeadTime(flagID uint, deadTime string) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("time", deadTime)
	return result.Error
}

// 更新用户状态
func UpdateUserStatus(id uint, status string) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("status", status)
	return result.Error
}

// 发布帖子
func AddPostToDB(Id uint, post model.Post) error {
	post.UserID = Id
	result := DB.Create(&post)
	return result.Error
}

// 删除帖子
func DeletePostFromDB(postID uint) error {
	// 先删除帖子的所有评论
	if err := DB.Where("post_id = ?", postID).Delete(&model.PostComment{}).Error; err != nil {
		return err
	}

	// 删除帖子的所有点赞记录
	if err := DB.Where("post_id = ?", postID).Delete(&model.UserPostLike{}).Error; err != nil {
		return err
	}

	// 最后删除帖子本身
	result := DB.Delete(&model.Post{}, postID)
	return result.Error
}

// 根据FlagID删除关联的帖子
func DeletePostsByFlagID(flagID uint) error {
	result := DB.Where("flag_id = ?", flagID).Delete(&model.Post{})
	return result.Error
}

// 添加评论
func AddPostCommentToDB(postId uint, comment model.PostComment) error {
	comment.PostID = postId
	result := DB.Create(&comment)
	return result.Error
}

// 删除评论
func DeletePostCommentFromDB(commentID uint) error {
	result := DB.Delete(&model.PostComment{}, commentID)
	return result.Error
}

// 根据关键词找帖子
func SearchPosts(keyword string) ([]model.Post, error) {
	var posts []model.Post
	like := "%" + keyword + "%"
	err := DB.Preload("User").Preload("Comments").
		Where("title LIKE ? OR content LIKE ?", like, like).Find(&posts).Error
	return posts, err
}

// 获取最近打卡的十个人
func GetRecentDoneFlags() ([]model.User, error) {
	var users []model.User
	result := DB.Where("had_done = ?", true).Order("do_flag desc").Limit(10).Find(&users)
	return users, result.Error
}

// 已完成的flag列表
func GetDoneFlagsByUserID(userID uint) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ? AND had_done = ?", userID, true).Find(&flags)
	return flags, result.Error
}

// 未完成的flag列表
func GetUndoneFlagsByUserID(userID uint) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ? AND had_done = ?", userID, false).Find(&flags)
	return flags, result.Error
}

// 工具函数，便于创造成就
func AddAchievementToDB(achievement model.Achievement) error {
	result := DB.Create(&achievement)
	return result.Error
}

// 插入单个成就（用于补全缺失成就）
func InsertAchievement(userID uint, name string, description string) error {
	achievement := model.Achievement{
		UserID:      userID,
		Name:        name,
		Description: description,
		HadDone:     false,
	}
	return DB.Create(&achievement).Error
}

// 批量插入成就
func BatchCreateAchievements(achievements []model.Achievement) error {
	if len(achievements) == 0 {
		return nil
	}
	return DB.Create(&achievements).Error
}

// 删除已使用的验证码（防止重复使用）
func DeleteEmailCode(email string) error {
	return DB.Where("email = ?", email).Delete(&model.EmailCode{}).Error
}

// 用户积分增加（原子操作，避免并发问题）
func CountAddDB(userID uint, count int) error {
	// 原子更新用户总积分
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&model.User{}).Where("id = ?", userID).Update("count", gorm.Expr("count + ?", count)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录积分变动日志，便于统计“今日获得积分”
	pl := model.PointsLog{
		UserID:    userID,
		Amount:    count,
		CreatedAt: time.Now(),
	}
	if err := tx.Create(&pl).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 获取今日获得的积分（按积分日志求和）
func GetTodayPoints(user_id uint) (int, error) {
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.Add(24 * time.Hour)

	var total struct{ Sum int }
	// 使用原生 SQL 聚合
	row := DB.Model(&model.PointsLog{}).Select("COALESCE(SUM(amount),0) as sum").Where("user_id = ? AND created_at >= ? AND created_at < ?", user_id, start, end).Scan(&total)
	if row.Error != nil {
		return 0, row.Error
	}
	return total.Sum, nil
}

// 用户flaga完成数量增加
func FlagNumberAddDB(userID uint, flagnumber int) error {
	result := DB.Model(&model.User{}).Where("id = ?", userID).Update("flag_number", flagnumber)
	return result.Error
}

// 获取所有用户，按积分排序
func GetUserByCount() ([]model.User, error) {
	var users []model.User
	result := DB.Order("count desc").Limit(20).Find(&users)
	return users, result.Error
}

// 获取所有20个用户，按月学习时间排序
func GetUserByMonthLearnTime() ([]model.User, error) {
	var users []model.User
	result := DB.Order("month_learn_time desc").Limit(20).Find(&users)
	return users, result.Error
}

// 获取20个用户，按总打卡数量排序
func GetUserByDaka() ([]model.User, error) {
	var users []model.User
	result := DB.Order("daka desc").Limit(20).Find(&users)
	return users, result.Error
}

// 20个用户按完成flag数量排序
func GetUserByFlagNumber() ([]model.User, error) {
	var users []model.User
	result := DB.Order("flag_number desc").Limit(20).Find(&users)
	return users, result.Error
}

// 通过flag id找到对应的flag
func GetFlagByID(flagID uint) (model.Flag, error) {
	var flag model.Flag
	result := DB.Where("id = ?", flagID).First(&flag)
	return flag, result.Error
}

// 获取所有的帖子（包含用户信息）
func GetAllPosts() ([]model.Post, error) {
	var posts []model.Post
	result := DB.Preload("Comments.User").Preload("User").Order("created_at desc").Find(&posts)
	return posts, result.Error
}

// 根据ID获取单个帖子
func GetPostByID(postID uint) (model.Post, error) {
	var post model.Post
	result := DB.Preload("Comments.User").Preload("User").First(&post, postID)
	return post, result.Error
}

// 根据ID获取单个评论
func GetCommentByID(commentID uint) (model.PostComment, error) {
	var comment model.PostComment
	result := DB.Preload("User").First(&comment, commentID)
	return comment, result.Error
}

// 获取所有可见的flag
func GetVisibleFlags() ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Preload("FlagComment").Where("is_public = ?", true).Find(&flags)
	return flags, result.Error
}

// 每天自动生成新的时间记录表
func AddNewLearnTimeToDB(user_id uint) error {
	err := DB.Create(&model.LearnTime{
		UserID:   user_id,
		Duration: 0,
	}).Error
	return err
}

// 更新学习时长
func UpdateLearnTimeDuration(user_id uint, duration int) error {
	var learnTime model.LearnTime
	// 🔧 修复：按当天日期查找/创建记录
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	err := DB.Where("user_id = ? AND created_at >= ? AND created_at < ?", user_id, todayStart, todayEnd).First(&learnTime).Error
	if err != nil {
		// 如果今天没有记录，创建新记录
		if err.Error() == "record not found" {
			learnTime = model.LearnTime{
				UserID:    user_id,
				Duration:  duration,
				CreatedAt: today,
			}
			return DB.Create(&learnTime).Error
		}
		return err
	}
	// 今天已有记录，累加时长
	learnTime.Duration += duration
	err = DB.Save(&learnTime).Error
	return err
}

// 获取今天的学习时长记录
func GetTodayLearnTime(user_id uint) (model.LearnTime, error) {
	var learnTime model.LearnTime
	// 🔧 修复：只查询当天的记录
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	err := DB.Where("user_id = ? AND created_at >= ? AND created_at < ?", user_id, todayStart, todayEnd).First(&learnTime).Error
	return learnTime, err
}

// 获取7天的学习时长（补全缺失日期）
func GetSevenDaysLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Find(&learnTime).Error
	if err != nil {
		return nil, err
	}

	// 创建日期映射（累加同一天的时长）
	dataMap := make(map[string]int)
	for _, record := range learnTime {
		dateStr := record.CreatedAt.Format("2006-01-02")
		if record.Duration >= 0 {
			dataMap[dateStr] += record.Duration
		}
	}

	// 补全最近7天的数据（从6天前到今天）
	result := make([]model.LearnTime, 7)
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -6+i) // 从6天前开始
		dateStr := date.Format("2006-01-02")
		duration := 0
		if val, ok := dataMap[dateStr]; ok {
			duration = val
		}
		result[i] = model.LearnTime{
			UserID:    user_id,
			CreatedAt: date,
			Duration:  duration,
		}
	}
	return result, nil
}

// 获取用户最近30天的学习时长记录（补全缺失日期）
func GetRecentLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Find(&learnTime).Error
	if err != nil {
		return nil, err
	}

	// 创建日期映射（只保存非负值）
	dataMap := make(map[string]int)
	for _, record := range learnTime {
		dateStr := record.CreatedAt.Format("2006-01-02")
		if record.Duration >= 0 {
			dataMap[dateStr] = record.Duration
		}
	}

	// 补全最近30天的数据（从29天前到今天）
	result := make([]model.LearnTime, 30)
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -29+i) // 从29天前开始
		dateStr := date.Format("2006-01-02")
		duration := 0
		if val, ok := dataMap[dateStr]; ok {
			duration = val
		}
		result[i] = model.LearnTime{
			UserID:    user_id,
			CreatedAt: date,
			Duration:  duration,
		}
	}
	return result, nil
}

// 获取用户最近180天的学习时长记录（补全缺失日期，返回20个数据点）
func GetRecent180LearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Find(&learnTime).Error
	if err != nil {
		return nil, err
	}

	// 创建日期映射（只保存非负值）
	dataMap := make(map[string]int)
	for _, record := range learnTime {
		dateStr := record.CreatedAt.Format("2006-01-02")
		if record.Duration >= 0 {
			dataMap[dateStr] = record.Duration
		}
	}

	// 生成20个数据点（覆盖180天，从最早到最晚）
	result := make([]model.LearnTime, 20)
	for i := 0; i < 20; i++ {
		// 每个数据点代表9天的聚合（180/20=9）
		// 从179天前开始，每9天一个点
		startDay := 179 - i*9
		date := time.Now().AddDate(0, 0, -startDay)

		// 聚合该数据点对应的9天数据（当前天及之前8天）
		totalDuration := 0
		for j := 0; j < 9; j++ {
			checkDate := date.AddDate(0, 0, -j)
			checkDateStr := checkDate.Format("2006-01-02")
			if val, ok := dataMap[checkDateStr]; ok {
				if val >= 0 {
					totalDuration += val
				}
			}
		}

		result[i] = model.LearnTime{
			UserID:    user_id,
			CreatedAt: date,
			Duration:  totalDuration,
		}
	}
	return result, nil
}

// 获取当前月份的学习时长记录（补全缺失日期）
func GetCurrentMonthLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Find(&learnTime).Error
	if err != nil {
		return nil, err
	}

	// 创建日期映射（只保存非负值）
	dataMap := make(map[string]int)
	for _, record := range learnTime {
		dateStr := record.CreatedAt.Format("2006-01-02")
		if record.Duration >= 0 {
			dataMap[dateStr] = record.Duration
		}
	}

	// 获取当前月份的天数
	now := time.Now()
	year, month, _ := now.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	daysInMonth := now.Day() // 从1号到今天

	// 补全当前月份的数据
	result := make([]model.LearnTime, daysInMonth)
	for i := 0; i < daysInMonth; i++ {
		date := firstDay.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		duration := 0
		if val, ok := dataMap[dateStr]; ok {
			duration = val
		}
		result[i] = model.LearnTime{
			UserID:    user_id,
			CreatedAt: date,
			Duration:  duration,
		}
	}
	return result, nil
}

// 获取最近6个月的学习时长记录（每月一个数据点）
func GetRecent6MonthsLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Find(&learnTime).Error
	if err != nil {
		return nil, err
	}

	// 创建日期映射（只保存非负值）
	dataMap := make(map[string]int)
	for _, record := range learnTime {
		dateStr := record.CreatedAt.Format("2006-01-02")
		if record.Duration >= 0 {
			dataMap[dateStr] = record.Duration
		}
	}

	// 生成6个月的数据点
	result := make([]model.LearnTime, 6)
	now := time.Now()

	for i := 0; i < 6; i++ {
		// 从5个月前到当前月
		targetMonth := now.AddDate(0, -5+i, 0)
		year, month, _ := targetMonth.Date()

		// 获取该月的第一天和最后一天
		firstDay := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
		lastDay := firstDay.AddDate(0, 1, -1)

		// 如果是当前月，只统计到今天
		if year == now.Year() && month == now.Month() {
			lastDay = now
		}

		// 聚合该月所有天的数据
		totalDuration := 0
		for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			if val, ok := dataMap[dateStr]; ok {
				if val >= 0 {
					totalDuration += val
				}
			}
		}

		// 代表日期始终为该月1号
		repDate := firstDay
		result[i] = model.LearnTime{
			UserID:    user_id,
			CreatedAt: repDate,
			Duration:  totalDuration,
		}
		fmt.Printf("6月聚合[%d]: %s, 时长: %d\n", i, repDate.Format("2006-01-02"), totalDuration)
	}
	return result, nil
}

// 存user
func SaveUserToDB(user model.User) error {
	result := DB.Save(&user)
	return result.Error
}

// 更新用户头像
func UpdateUserHeadShow(userID uint, headShow int) error {
	result := DB.Model(&model.User{}).Where("id = ?", userID).Update("head_show", headShow)
	return result.Error
}

// 获取所有用户
func GetAllUser() ([]model.User, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}
	var users []model.User
	// 只取每个邮箱最新一条（假设id自增，取最大id）
	result := DB.Raw(`
		   SELECT * FROM users u
		   WHERE u.id = (
			   SELECT MAX(id) FROM users WHERE email = u.email
		   )
	   `).Scan(&users)
	return users, result.Error
}

// 完成成就
func UpdateAchievementHadDone(usrID uint, name string) error {
	result := DB.Model(&model.Achievement{}).Where("name=?", name).Where("user_id=?", usrID).Update("had_done", true)
	return result.Error
}

// 获取用户成就列表
func GetAchievementsByUserID(userID uint) ([]model.Achievement, error) {
	var achievements []model.Achievement
	result := DB.Where("user_id = ?", userID).Find(&achievements)
	return achievements, result.Error
}

// 根据成就名使它完成
func GetAchievementByName(usrID uint, name string) (model.Achievement, error) {
	var achievement model.Achievement
	result := DB.Where("name=? AND user_id=?", name, usrID).First(&achievement)
	return achievement, result.Error
}

// 计算连续打卡天数（用于积分奖励计算）
func calculateConsecutiveDays(userID uint) int {
	var records []model.Daka_number
	// 获取最近30天的打卡记录，按日期倒序
	err := DB.Where("user_id = ? AND had_done = ?", userID, true).
		Order("daka_date desc").
		Limit(30).
		Find(&records).Error

	if err != nil || len(records) == 0 {
		return 1 // 默认返回1（今天是第一天）
	}

	consecutive := 1
	today := time.Now().Format("2006-01-02")

	for i := 0; i < len(records); i++ {
		recordDate := records[i].DaKaDate.Format("2006-01-02")

		if i == 0 {
			// 第一条记录应该是今天或昨天
			if recordDate != today {
				dayDiff := daysBetween(records[i].DaKaDate, time.Now())
				if dayDiff > 1 {
					return 1 // 中断了，重新开始计数
				}
			}
			continue
		}

		// 检查与前一天的记录是否连续
		prevDate := records[i-1].DaKaDate
		dayDiff := daysBetween(records[i].DaKaDate, prevDate)

		if dayDiff == 1 {
			consecutive++
		} else {
			break // 不连续，停止计数
		}
	}

	return consecutive
}

// 计算两个日期之间的天数差
func daysBetween(date1, date2 time.Time) int {
	d1 := time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, time.UTC)
	d2 := time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, time.UTC)
	return int(d2.Sub(d1).Hours() / 24)
}

// 添加打卡记录
func DakaNumberToDB(user_id uint) error {
	// 先查询是否存在打卡记录
	var dakaNumber model.Daka_number
	err := DB.Where("user_id = ?", user_id).Order("id desc").First(&dakaNumber).Error

	if err == gorm.ErrRecordNotFound {
		// 如果不存在,创建新的打卡记录并设置为已打卡
		err := DB.Create(&model.Daka_number{
			UserID:    user_id,
			HadDone:   true,
			DaKaDate:  time.Now(),
			MonthDaka: 1, // 第一次打卡，月打卡数为1
		}).Error
		if err != nil {
			return err
		}
		// 更新用户总打卡数
		err = DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka + ?", 1)).Error
		if err != nil {
			return err
		}

		// 🔧 添加打卡积分：基础20分（第一次打卡）
		return CountAddDB(user_id, 20)
	}

	if err != nil {
		return err
	}

	// 检查今天是否已经打卡
	today := time.Now().Format("2006-01-02")
	recordDate := dakaNumber.DaKaDate.Format("2006-01-02")

	if recordDate == today {
		// 今天已经打卡，切换状态（支持取消打卡）
		newStatus := !dakaNumber.HadDone
		err = DB.Model(&model.Daka_number{}).Where("id = ?", dakaNumber.ID).Update("had_done", newStatus).Error
		if err != nil {
			return err
		}
		// 更新用户总打卡数（取消打卡则-1，打卡则+1）
		if newStatus {
			// 打卡：增加打卡数并添加积分
			err = DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka + ?", 1)).Error
			if err != nil {
				return err
			}
			// 🔧 计算连续打卡积分：基础20分 + 连续奖励
			consecutiveDays := calculateConsecutiveDays(user_id)
			points := 20
			if consecutiveDays >= 10 {
				points += 20 // 连续10天：额外+10+10=20，总老40分/天
			} else if consecutiveDays >= 5 {
				points += 10 // 连续5天：额外+10，总老30分/天
			}
			return CountAddDB(user_id, points)
		} else {
			// 取消打卡：减少打卡数，不扣积分（已获得的积分保留）
			return DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka - ?", 1)).Error
		}
	} else {
		// 不是今天的记录，创建新的打卡记录
		err := DB.Create(&model.Daka_number{
			UserID:    user_id,
			HadDone:   true,
			DaKaDate:  time.Now(),
			MonthDaka: dakaNumber.MonthDaka + 1, // 月打卡数+1
		}).Error
		if err != nil {
			return err
		}
		// 更新用户总打卡数
		err = DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka + ?", 1)).Error
		if err != nil {
			return err
		}

		// 🔧 计算连续打卡积分：基础20分 + 连续奖励
		consecutiveDays := calculateConsecutiveDays(user_id)
		points := 20
		if consecutiveDays >= 10 {
			points += 20 // 连续10天：额外+10+10=20，总老40分/天
		} else if consecutiveDays >= 5 {
			points += 10 // 连续5天：额外+10，总老30分/天
		}
		return CountAddDB(user_id, points)
	}
}

// 添加打卡记录
func AddDakaNumberToDB(user_id uint) error {
	// 先查询是否存在打卡记录
	var dakaNumber model.Daka_number
	err := DB.Where("user_id=?", user_id).Order("id desc").First(&dakaNumber).Error

	if err == gorm.ErrRecordNotFound {
		// 如果不存在,创建新的打卡记录
		return AddNewDakaNumberToDB(user_id)
	}

	if err != nil {
		return err
	}

	// 如果存在,更新monthDaka
	err = DB.Model(&model.Daka_number{}).Where("user_id=?", user_id).Order("id desc").Limit(1).Update("monthDaka", gorm.Expr("monthDaka + ?", 1)).Error
	return err
}

// 获取用户最近的打卡记录
func GetRecentDakaNumber(user_id uint) (model.Daka_number, error) {
	var daka_number model.Daka_number
	err := DB.Where("user_id = ?", user_id).Order("id desc").First(&daka_number).Error
	return daka_number, err
}

// 获取用户本月所有打卡记录
func GetMonthDakaRecords(user_id uint) ([]model.Daka_number, error) {
	var records []model.Daka_number
	// 获取本月第一天
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := DB.Where("user_id = ? AND had_done = true AND daka_date >= ?", user_id, firstDay).
		Order("daka_date asc").
		Find(&records).Error
	return records, err
}

// 每日更新打卡状态
func UpdateDakaHadDone(userid uint) error {
	result := DB.Model(&model.Daka_number{}).Where("user_id = ?", userid).Update("had_done", false)
	return result.Error
}

// 每月建立打卡记录
func AddNewDakaNumberToDB(user_id uint) error {
	err := DB.Create(&model.Daka_number{
		UserID:    user_id,
		HadDone:   false,
		DaKaDate:  time.Now(),
		MonthDaka: 0,
	}).Error
	return err
}

// 存验证码
func SaveEmailCodeToDB(code string, email string) error {
	var emailCode model.EmailCode
	emailCode.Code = code
	emailCode.Email = email
	emailCode.CreatedAt = time.Now()
	emailCode.Expires = time.Now().Add(time.Minute * 5) // 设置过期时间为5分钟后
	result := DB.Create(&emailCode)
	return result.Error
}

// 根据邮箱找到第一个验证码
func GetEmailCodeByEmail(email string) (model.EmailCode, error) {
	var emailCode model.EmailCode
	result := DB.Where("email = ?", email).Order("created_at desc").First(&emailCode)
	return emailCode, result.Error
}

// 删除过期的验证码
func DeleteExpiredEmailCodes() error {
	result := DB.Where("expires < ?", time.Now()).Delete(&model.EmailCode{})
	return result.Error
}

// 🔒 安全加固：检查邮箱验证码发送频率限制（5分钟1次 + 每天最多5次）
func CheckEmailCodeRateLimit(email string) (bool, time.Time, error) {
	var emailCode model.EmailCode

	// 检查5分钟内是否发送过
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	err := DB.Where("email = ? AND created_at > ?", email, fiveMinutesAgo).Order("created_at desc").First(&emailCode).Error
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			// 找到了最近5分钟的记录，不能发送
			return false, emailCode.CreatedAt, nil
		}
		// 数据库错误
		return false, time.Time{}, err
	}

	// 检查今天发送次数（不超过5次）
	todayStart := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	DB.Model(&model.EmailCode{}).Where("email = ? AND created_at > ?", email, todayStart).Count(&todayCount)
	if todayCount >= 5 {
		return false, emailCode.CreatedAt, fmt.Errorf("今日验证码发送次数已达上限")
	}

	// 可以发送
	return true, time.Time{}, nil
}

// 修改用户的验证状态
func UpdateUserExistStatus(email string) error {
	result := DB.Model(&model.User{}).Where("email = ?", email).Update("exist", true)
	return result.Error
}

// 存储用户提醒时间
func UpdateUserRemindTime(id uint, hour int, min int) error {
	// 同时更新兼容字段和新的学习提醒字段，保持向后兼容
	result := DB.Model(&model.User{}).Where("id=?", id).Updates(map[string]interface{}{"remind_hour": hour, "remind_min": min, "study_remind_hour": hour, "study_remind_min": min})
	return result.Error
}

// 是否开启提醒
func UpdateUserRemindStatus(id uint, IsRemind bool) error {
	// 更新兼容字段和新的学习提醒总开关
	result := DB.Model(&model.User{}).Where("id=?", id).Updates(map[string]interface{}{"is_remind": IsRemind, "is_study_remind": IsRemind})
	return result.Error
}

// 更新学习提醒开关（新）
func UpdateUserStudyRemindStatus(id uint, enabled bool) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("is_study_remind", enabled)
	return result.Error
}

// 更新 Flag 提醒开关（用户级别的总开关）
func UpdateUserFlagRemindStatus(id uint, enabled bool) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("is_flag_remind", enabled)
	return result.Error
}

// flag点赞
func UpdateFlagLikes(flagID uint, like int) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("likes", like)
	return result.Error
}

// post点赞
// 切换帖子点赞状态（自动判断点赞/取消点赞）- 使用事务确保原子性
func TogglePostLike(postID uint, userID uint) (int, error) {
	utils.LogInfo("TogglePostLike 函数被调用", map[string]interface{}{
		"post_id": postID,
		"user_id": userID,
	})

	// 使用事务确保原子性
	tx := DB.Begin()
	if tx.Error != nil {
		utils.LogError("开启事务失败", map[string]interface{}{
			"post_id": postID,
			"user_id": userID,
			"error":   tx.Error.Error(),
		})
		return 0, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.LogError("事务执行中发生panic", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"panic":   r,
			})
		}
	}()

	// 1. 检查是否已点赞
	var like model.UserPostLike
	err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error

	if err == nil {
		// 已点赞，取消点赞
		if err := tx.Delete(&like).Error; err != nil {
			tx.Rollback()
			utils.LogError("取消点赞失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		// 减少点赞数，确保不会小于0
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).Update("like", gorm.Expr("CASE WHEN `like` > 0 THEN `like` - 1 ELSE 0 END")).Error; err != nil {
			tx.Rollback()
			utils.LogError("更新点赞数失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		// 获取更新后的点赞数
		var post model.Post
		if err := tx.Where("id = ?", postID).First(&post).Error; err != nil {
			tx.Rollback()
			utils.LogError("获取更新后点赞数失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			utils.LogError("提交事务失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		utils.LogInfo("取消点赞成功", map[string]interface{}{
			"post_id":   postID,
			"user_id":   userID,
			"new_likes": post.Like,
		})
		return post.Like, nil

	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 未点赞，添加点赞
		newLike := model.UserPostLike{
			UserID:    userID,
			PostID:    postID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&newLike).Error; err != nil {
			tx.Rollback()
			utils.LogError("点赞失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		// 增加点赞数
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).Update("like", gorm.Expr("`like` + 1")).Error; err != nil {
			tx.Rollback()
			utils.LogError("更新点赞数失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		// 获取更新后的点赞数
		var post model.Post
		if err := tx.Where("id = ?", postID).First(&post).Error; err != nil {
			tx.Rollback()
			utils.LogError("获取更新后点赞数失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			utils.LogError("提交事务失败", map[string]interface{}{
				"post_id": postID,
				"user_id": userID,
				"error":   err.Error(),
			})
			return 0, err
		}

		utils.LogInfo("点赞成功", map[string]interface{}{
			"post_id":   postID,
			"user_id":   userID,
			"new_likes": post.Like,
		})
		return post.Like, nil

	} else {
		// 其他数据库错误
		tx.Rollback()
		utils.LogError("查询点赞状态失败", map[string]interface{}{
			"post_id": postID,
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, err
	}
}

func UpdatePostLikes(postID uint, like int) error {
	result := DB.Model(&model.Post{}).Where("id = ?", postID).Update("like", like)
	return result.Error
}

// 获取帖子点赞数
func GetFlagLikes(flagID uint) (int, error) {
	var flag model.Flag
	result := DB.Where("id = ?", flagID).First(&flag)
	return flag.Likes, result.Error
}

// 获取帖子点赞
func GetPostLikes(flagID uint) (int, error) {
	var post model.Post
	result := DB.Where("id = ?", flagID).First(&post)
	return post.Like, result.Error
}

// 获取用户点过赞的帖子ID列表
func GetLikedPostIDsByUser(userID uint) ([]uint, error) {
	var likes []model.UserPostLike
	if err := DB.Where("user_id = ?", userID).Find(&likes).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(likes))
	for _, l := range likes {
		ids = append(ids, l.PostID)
	}
	return ids, nil
}

// 储存标签
func SaveLabelToDB(id uint, labal string) error {
	err := DB.Model(&model.Label{}).Where("user_id = ?", id).Update(labal, gorm.Expr(labal+" + ?", 1)).Error
	return err
}

// 调取用户不同种类的标签数
func GetLabelByUserID(userID uint) (model.Label, error) {
	var label model.Label
	err := DB.Where("user_id = ?", userID).First(&label).Error
	// 如果用户没有标签记录，创建一个默认的
	if err != nil {
		if err.Error() == "record not found" {
			label = model.Label{
				UserID: userID,
				Life:   0,
				Study:  0,
				Work:   0,
				Like:   0,
				Sport:  0,
			}
			// 创建默认记录
			DB.Create(&label)
			return label, nil
		}
		return label, err
	}
	return label, nil
}

// 存储埋点
func AddTrackPointToDB(user_id uint, event string) error {
	var trackPoint model.TrackPoint
	trackPoint.UserID = user_id
	trackPoint.Event = event
	trackPoint.Timestamp = time.Now()
	result := DB.Create(&trackPoint)
	return result.Error
}

// 按时间读取所有埋点
func GetTrackPointsByUserIDAndTime() ([]model.TrackPoint, error) {
	var trackPoints []model.TrackPoint
	err := DB.Order("timestam desc").Find(&trackPoints).Error
	return trackPoints, err
}

// 自从数据库中删除验证码
func DeleteEmailCodeByEmail(email string) error {
	result := DB.Where("email = ?", email).Delete(&model.EmailCode{})
	return result.Error
}

// 保存聊天消息
func SaveChatMessage(message *model.ChatMessage) error {
	result := DB.Create(message)
	return result.Error
}

// 获取谈玄斋历史消息（最近30条）
func GetChatHistory(roomID string, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := DB.Preload("User").Where("room_id = ?", roomID).Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	// 反转顺序，让最早的消息在前面
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// 获取私聊历史消息（最近30条）
func GetPrivateChatHistory(userID1, userID2 uint, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := DB.Preload("User").
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", userID1, userID2, userID2, userID1).
		Order("created_at desc").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	// 反转顺序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// Conversation 会话信息
type Conversation struct {
	UserID        uint      `json:"user_id"`
	UserName      string    `json:"user_name"`
	UserAvatar    string    `json:"user_avatar"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
}

// 获取私聊会话列表（按最后消息时间排序）
func GetPrivateConversations(userID uint) ([]Conversation, error) {
	var conversations []Conversation

	// 简化版本：直接查询所有私聊消息，在Go中处理分组
	var messages []model.ChatMessage
	err := DB.Preload("User").
		Where("(from_user_id = ? OR to_user_id = ?) AND (room_id = '' OR room_id IS NULL)", userID, userID).
		Order("created_at DESC").
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 按对方用户ID分组，保留最新消息
	conversationMap := make(map[uint]*Conversation)
	for _, msg := range messages {
		// 确定对方用户ID
		var otherUserID uint
		if msg.FromUserID == userID {
			otherUserID = msg.ToUserID
		} else {
			otherUserID = msg.FromUserID
		}

		// 如果已存在且不是更新的消息，跳过
		if existing, exists := conversationMap[otherUserID]; exists {
			if !msg.CreatedAt.After(existing.LastMessageAt) {
				continue
			}
		}

		// 获取用户信息
		var user model.User
		if err := DB.First(&user, otherUserID).Error; err != nil {
			continue
		}

		// 构建头像路径（使用 utils.GetAvatarPath 统一返回 /api/avatar/:id）
		var avatar string
		if user.HeadShow > 0 {
			avatar = utils.GetAvatarPath(user.HeadShow)
		}

		conversationMap[otherUserID] = &Conversation{
			UserID:        user.ID,
			UserName:      user.Name,
			UserAvatar:    avatar,
			LastMessage:   msg.Content,
			LastMessageAt: msg.CreatedAt,
			UnreadCount:   0, // TODO: 实现未读计数
		}
	}

	// 转换为切片并按时间排序
	for _, conv := range conversationMap {
		conversations = append(conversations, *conv)
	}

	// 按最后消息时间排序（最新的在前）
	for i := 0; i < len(conversations); i++ {
		for j := i + 1; j < len(conversations); j++ {
			if conversations[j].LastMessageAt.After(conversations[i].LastMessageAt) {
				conversations[i], conversations[j] = conversations[j], conversations[i]
			}
		}
	}

	return conversations, nil
}

// 每天凌晨4点：将所有用户当天的学习计时置为无效（不计入学习时长）
func InvalidateAllTodayLearnTime() error {
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	// 批量更新：将今天所有学习时长置为-1（或可加 is_valid 字段，现用-1表示无效）
	err := DB.Model(&model.LearnTime{}).
		Where("created_at >= ? AND created_at < ? AND duration > 0", todayStart, todayEnd).
		Update("duration", -1).Error
	return err
}
