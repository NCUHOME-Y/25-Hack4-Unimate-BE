package repository

import (
	"os"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model" // 你的自定义包
	"github.com/sirupsen/logrus"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

// 链接数据库
func DBconnect() {
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.LogError("数据库连接失败", logrus.Fields{"error": err})
		return
	}
	DB = db
	DB.AutoMigrate(&model.User{}, &model.Flag{}, &model.Post{}, &model.PostComment{}, &model.Achievement{}, &model.LearnTime{}, &model.Daka_number{}, &model.EmailCode{}, &model.FlagComment{}, &model.TrackPoint{}, &model.ChatMessage{})
}

// user添加到数据库
func AddUserToDB(user model.User) error {
	result := DB.Create(&user)
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
	result := DB.Where("user_id = ?", userID).Order("priority").Find(&flags)
	return flags, result.Error
}

// 通过用户邮箱获取用户
func GetUserByEmail(Email string) (model.User, error) {
	var user model.User
	result := DB.Preload("Achievements").Preload("Flags").Preload("Posts").Where("email = ?", Email).First(&user)
	return user, result.Error
}

// 通过用户名获取用户
func GetUserByName(name string) (model.User, error) {
	var user model.User
	result := DB.Where("name = ?", name).First(&user)
	return user, result.Error
}

// 通过用户ID获取用户
func GetUserByID(userID uint) (model.User, error) {
	var user model.User
	result := DB.Preload("Achievements").Preload("Flags").Preload("Posts").Where("id = ?", userID).First(&user)
	return user, result.Error
}

// 搜索关键词查询用户，可以是邮箱是用户名
func SearchUsers(keyword string) ([]model.User, error) {
	var users []model.User
	like := "%" + keyword + "%"
	err := DB.Preload("Achievement").
		Preload("Flags").
		Preload("Posts").
		Where("user_name=like AND user_email=like", like, like).Find(&users).Error // 把 Flags 一起查出来
	return users, err
}

// 更新flag的可见性
func UpdateFlagVisibility(flagID uint, isHidden bool) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("is_hiden", isHidden)
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
	result := DB.Model(&model.User{}).Where("id=?", id).Update("Name", newName)
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
	result := DB.Delete(&model.Post{}, postID)
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

// 用户积分增加
func CountAddDB(userID uint, count int) error {
	result := DB.Model(&model.User{}).Where("id = ?", userID).Update("count", count)
	return result.Error
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
	result := DB.Preload("FlagComment").Where("is_hiden = ?", false).Find(&flags)
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
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").First(&learnTime).Error
	if err != nil {
		return err
	}
	learnTime.Duration += duration
	err = DB.Save(&learnTime).Error
	return err
}

// 获取今天的学习时长记录
func GetTodayLearnTime(user_id uint) (model.LearnTime, error) {
	var learnTime model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Limit(1).First(&learnTime).Error
	return learnTime, err
}

// 获取7天的学习时长
func GetSevenDaysLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Limit(7).Find(&learnTime).Error
	return learnTime, err
}

// 获取用户最近30天的学习时长记录
func GetRecentLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Limit(30).Find(&learnTime).Error
	return learnTime, err
}

// 获取用户最近180天的学习时长记录
func GetRecent180LearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Limit(180).Find(&learnTime).Error
	return learnTime, err
}

// 存user
func SaveUserToDB(user model.User) error {
	result := DB.Save(&user)
	return result.Error
}

// 获取所有用户
func GetAllUser() ([]model.User, error) {
	var users []model.User
	result := DB.Find(&users)
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
		return DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka + ?", 1)).Error
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
			return DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka + ?", 1)).Error
		} else {
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
		return DB.Model(&model.User{}).Where("id = ?", user_id).Update("daka", gorm.Expr("daka + ?", 1)).Error
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

// 检查邮箱最近1分钟内是否发送过验证码
func CheckEmailCodeRateLimit(email string) (bool, time.Time, error) {
	var emailCode model.EmailCode
	oneMinuteAgo := time.Now().Add(-time.Minute)
	err := DB.Where("email = ? AND created_at > ?", email, oneMinuteAgo).Order("created_at desc").First(&emailCode).Error
	if err == gorm.ErrRecordNotFound {
		// 没有找到最近1分钟的记录，可以发送
		return true, time.Time{}, nil
	}
	if err != nil {
		// 数据库错误
		return false, time.Time{}, err
	}
	// 找到了最近的记录，不能发送，返回创建时间
	return false, emailCode.CreatedAt, nil
}

// 修改用户的验证状态
func UpdateUserExistStatus(email string) error {
	result := DB.Model(&model.User{}).Where("email = ?", email).Update("exist", true)
	return result.Error
}

// 存储用户提醒时间
func UpdateUserRemindTime(id uint, hour int, min int) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Updates(map[string]interface{}{"remind_hour": hour, "remind_min": min})
	return result.Error
}

// 是否开启提醒
func UpdateUserRemindStatus(id uint, IsRemind bool) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("is_remind", IsRemind)
	return result.Error
}

// flag点赞
func UpdateFlagLikes(flagID uint, like int) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("likes", like)
	return result.Error
}

// post点赞
// 切换帖子点赞状态（自动判断点赞/取消点赞）
func TogglePostLike(postID uint, userID uint) (int, error) {
	utils.LogInfo("TogglePostLike 函数被调用", map[string]interface{}{
		"post_id": postID,
		"user_id": userID,
	})

	var post model.Post

	// 获取帖子当前点赞数
	if err := DB.Where("id = ?", postID).First(&post).Error; err != nil {
		utils.LogError("查询帖子失败", map[string]interface{}{
			"post_id": postID,
			"error":   err.Error(),
		})
		return 0, err
	}

	utils.LogInfo("查询到帖子", map[string]interface{}{
		"post_id":      postID,
		"current_like": post.Like,
	})

	// TODO: 实现用户点赞关系表来记录谁点赞了哪些帖子
	// 目前简化实现：直接增加点赞数
	// 生产环境应该：
	// 1. 检查 user_post_likes 表是否存在该用户对该帖子的点赞记录
	// 2. 如果存在则删除记录并减少点赞数
	// 3. 如果不存在则创建记录并增加点赞数

	// 简化实现：每次调用都增加点赞数（前端控制）
	newLikeCount := post.Like + 1

	utils.LogInfo("准备更新点赞数", map[string]interface{}{
		"post_id":  postID,
		"old_like": post.Like,
		"new_like": newLikeCount,
	})

	if err := DB.Model(&model.Post{}).Where("id = ?", postID).Update("like", newLikeCount).Error; err != nil {
		utils.LogError("更新点赞数失败", map[string]interface{}{
			"post_id": postID,
			"error":   err.Error(),
		})
		return 0, err
	}

	utils.LogInfo("点赞数更新成功", map[string]interface{}{
		"post_id":  postID,
		"new_like": newLikeCount,
	})

	return newLikeCount, nil
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

// 获取聊天室历史消息（最近30条）
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

		// 构建头像路径
		var avatar string
		if user.HeadShow > 0 && user.HeadShow <= 6 {
			avatarFiles := []string{"131601", "131629", "131937", "131951", "132014", "133459"}
			avatar = "/src/assets/images/screenshot_20251114_" + avatarFiles[user.HeadShow-1] + ".png"
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
