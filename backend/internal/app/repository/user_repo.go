package repository

import (
	"Heckweek/internal/app/model" // 你的自定义包
	"os"

	"log"

	"github.com/joho/godotenv"

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
		log.Fatalf("Error loading_data .env file")
	}
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		return
	}
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
		return
	}
	DB = db
	DB.AutoMigrate(&model.User{}, &model.Flag{}, &model.Post{}, &model.PostComment{})
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

// 从数据库删除flag
func DeleteFlagFromDB(flagID uint) error {
	result := DB.Delete(&model.Flag{}, flagID)
	return result.Error
}

// 通过用户ID获取flag列表
func GetFlagsByUserID(userID uint) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ?", userID).Find(&flags)
	return flags, result.Error
}

// 通过用户邮箱获取用户
func GetUserByEmail(Email string) (model.User, error) {
	var user model.User
	DB.Preload("Flags").First(&user, Email)
	DB.Preload("Posts").First(&user, Email)
	result := DB.Where("email=?", Email).First(&user)
	return user, result.Error
}

// 通过用户ID获取用户
func GetUserByID(userID uint) (model.User, error) {
	var user model.User
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
func UpdateFlagHadDone(flagID uint, hadDone bool) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("had_done", hadDone)
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
