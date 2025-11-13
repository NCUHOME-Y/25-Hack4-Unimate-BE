package repository

import (
	"os"

	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model" // 你的自定义包
	"github.com/sirupsen/logrus"

	"github.com/joho/godotenv"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

// 链接数据库
func DBconnect() {
	err := godotenv.Load(".env")
	if err != nil {
		utils.LogError("加载 .env 文件失败", logrus.Fields{"error": err})
		return
	}
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.LogError("数据库连接失败", logrus.Fields{"error": err})
		return
	}
	DB = db
	DB.AutoMigrate(&model.User{}, &model.Flag{}, &model.Post{}, &model.Comment{}, &model.Achievement{}, &model.LearnTime{}, &model.Daka_number{}, &model.EmailCode{})
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

// 通过邮箱删除用户
func DeleteUserByEmail(email string) error {
	result := DB.Where("email = ?", email).Delete(&model.User{})
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
	DB.Preload("Achievement").First(&user, Email)
	DB.Preload("Flags").First(&user, Email)
	DB.Preload("Posts").First(&user, Email)
	result := DB.Where("email=?", Email).First(&user)
	return user, result.Error
}

// 通过用户ID获取用户
func GetUserByID(userID uint) (model.User, error) {
	var user model.User
	DB.Preload("Achievement").First(&user, userID)
	DB.Preload("Flags").First(&user, userID)
	DB.Preload("Posts").First(&user, userID) // 把 Flags 一起查出来
	result := DB.Where("id=?", userID).First(&user)
	return user, result.Error
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

func AddPostCommentToDB(postId uint, comment model.Comment) error {
	comment.CommentID = postId
	result := DB.Create(&comment)
	return result.Error
}

// 删除评论
func DeletePostCommentFromDB(commentID uint) error {
	result := DB.Delete(&model.Comment{}, commentID)
	return result.Error
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

// 通过flag id找到对应的flag
func GetFlagByID(flagID uint) (model.Flag, error) {
	var flag model.Flag
	result := DB.Where("id = ?", flagID).First(&flag)
	return flag, result.Error
}

// 获取所有的帖子
func GetAllPosts() ([]model.Post, error) {
	var posts []model.Post
	result := DB.Preload("Comments").Order("created_at desc").Find(&posts)
	return posts, result.Error
}

// 获取所有可见的flag
func GetVisibleFlags() ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("is_hiden = ?", false).Find(&flags)
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

// 获取用户最近的学习时长记录
func GetRecentLearnTime(user_id uint) ([]model.LearnTime, error) {
	var learnTime []model.LearnTime
	err := DB.Where("user_id = ?", user_id).Order("created_at desc").Limit(30).Find(&learnTime).Error
	return learnTime, err
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
	result := DB.Model(&model.Daka_number{}).Where("user_id = ?", user_id).Order("daka_date desc").Limit(1).Update("had_done", true)
	return result.Error
}

// 添加打卡记录
func AddDakaNumberToDB(user_id uint) error {
	err := DB.Model(&model.Daka_number{}).Where("user_id=?", user_id).Order("id desc").Limit(1).Update("monthDaka", gorm.Expr("monthDaka + ?", 1)).Error
	return err
}

// 获取用户最近的打卡记录
func GetRecentDakaNumber(user_id uint) (model.Daka_number, error) {
	var daka_number model.Daka_number
	err := DB.Where("user_id = ?", user_id).Order("daka_date desc").First(&daka_number).Error
	return daka_number, err
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

// 修改用户的验证状态
func UpdateUserExistStatus(email string) error {
	result := DB.Model(&model.User{}).Where("email = ?", email).Update("exist", true)
	return result.Error
}
