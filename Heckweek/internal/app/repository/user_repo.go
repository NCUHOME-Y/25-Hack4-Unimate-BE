package repository

import (
	"Number4_Heckweek_BE/Heckweek/internal/app/model" // 你的自定义包
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	users []model.User
	flags []model.Flag
)

func init() {
	err := godotenv.Load()
	if err != nil {
		zap.L().Error("Error loading .env file", zap.Error(err))
	}
}

// 链接数据库
func DBconnect() {
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.L().Error("Failed to connect to database", zap.Error(err))
	}
	DB = db
	DB.AutoMigrate(&model.User{}, &model.Flag{})
}

// flag添加到数据库
func addUserToDB(user model.User) error {
	result := DB.Create(&user)
	return result.Error
}

// flag添加到数据库
func addFlagToDB(flag model.Flag) error {
	result := DB.Create(&flag)
	return result.Error
}

// 从数据库删除flag
func deleteFlagFromDB(flagID uint) error {
	result := DB.Delete(&model.Flag{}, flagID)
	return result.Error
}

// 通过用户ID获取flag列表
func getFlagsByUserID(userID uint) ([]model.Flag, error) {
	var flags []model.Flag
	result := DB.Where("user_id = ?", userID).Find(&flags)
	return flags, result.Error
}

// 通过用户名获取用户
func getUserByName(name string) (model.User, error) {
	var user model.User
	result := DB.Where("name=?", name).First(&user)
	return user, result.Error
}
func getUserByID(userID uint) (model.User, error) {
	var user model.User
	result := DB.Where("id=?", userID).First(&user)
	return user, result.Error
}
func updateFlagVisibility(flagID uint, isHidden bool) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("is_hiden", isHidden)
	return result.Error
}
func updateFlagContent(flagID uint, newContent string) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("flag", newContent)
	return result.Error
}
func updatePlanContent(flagID uint, newPlanContent string) error {
	result := DB.Model(&model.Flag{}).Where("id = ?", flagID).Update("plan_content", newPlanContent)
	return result.Error
}
func updatePassword(id uint, newPassword string) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("Password", newPassword)
	return result.Error
}
