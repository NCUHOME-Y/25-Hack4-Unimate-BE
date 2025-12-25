package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetUserDisplayName 获取用户显示名称，如果用户名为空则返回默认值
func GetUserDisplayName(user model.User) string {
	if user.Name != "" {
		return user.Name
	}
	return "用户"
}

// GetUserDisplayNameByID 根据用户ID获取显示名称
func GetUserDisplayNameByID(userID uint) string {
	user, err := repository.GetUserByID(userID)
	if err != nil || user.Name == "" {
		return "用户"
	}
	return user.Name
}

// GenerateAndSendVerificationCode 生成并发送验证码，返回错误信息
func GenerateAndSendVerificationCode(email string) (string, error) {
	code := utils.GenerateCode()
	err := SendVerificationCodeEmail(email, code)
	if err != nil {
		utils.LogError("验证码邮件发送失败", logrus.Fields{"email": email, "error": err.Error()})
		return "", err
	}

	// 保存验证码到数据库
	repository.SaveEmailCodeToDB(code, email)
	utils.LogInfo("验证码已发送", logrus.Fields{"email": email})
	return code, nil
}

// GenerateAndSendIdentityCode 生成并发送身份验证码
func GenerateAndSendIdentityCode(email string) (string, error) {
	code := utils.GenerateCode()
	err := SendIdentityVerificationEmail(email, code)
	if err != nil {
		utils.LogError("身份验证码邮件发送失败", logrus.Fields{"email": email, "error": err.Error()})
		return "", err
	}

	// 保存验证码到数据库
	repository.SaveEmailCodeToDB(code, email)
	utils.LogInfo("身份验证码已发送", logrus.Fields{"email": email})
	return code, nil
}

// HandleBindError 处理请求参数绑定错误，统一返回400响应
func HandleBindError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return true
	}
	return false
}
