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
	DB    *gorm.DB
	Users []model.User
	Flags []model.Flag
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
	}
	DB = db
	DB.AutoMigrate(&model.User{}, &model.Flag{})
}

// flag添加到数据库
func AddUserToDB(user model.User) error {
	result := DB.Create(&user)
	return result.Error
}

// flag添加到数据库
func AddFlagToDB(Id uint, flag []model.Flag) error {
	result := DB.Model(&model.Flag{}).Where("user_id=?", Id).Update("flag", flag)
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
	result := DB.Where("email=?", Email).First(&user)
	return user, result.Error
}

// 通过用户ID获取用户
func GetUserByID(userID uint) (model.User, error) {
	var user model.User
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
func UpdateUserName(id uint, newName string) error {
	result := DB.Model(&model.User{}).Where("id=?", id).Update("Name", newName)
	return result.Error
}
