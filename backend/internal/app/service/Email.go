package service

import (
	"time"

	"log"
	"net/http"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 验证邮箱
func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			utils.LogError("绑定邮箱请求参数错误", nil)
			return
		}
		email, err := repository.GetEmailCodeByEmail(req.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取验证码失败,请重新再试..."})
			utils.LogError("获取邮箱验证码失败", nil)
			return
		}
		if email.Code != req.Code {
			c.JSON(400, gin.H{"error": "验证码错误"})
			utils.LogError("邮箱验证码错误", nil)
			return
		}
		if email.Expires.Before(time.Now()) {
			c.JSON(400, gin.H{"error": "验证码已过期"})
			utils.LogError("邮箱验证码已过期", nil)
			return
		}
		repository.UpdateUserExistStatus(req.Email)
		utils.LogInfo("邮箱验证成功", logrus.Fields{"user_email": req.Email})
		utils.SentEmail(req.Email, "邮箱验证成功", "恭喜您成功验证账户")
		c.JSON(200, gin.H{"success": true})
	}
}

func ForgetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "注册失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		user_exist, _ := repository.GetUserByEmail(user.Email)
		if user_exist.ID != 0 {
			c.JSON(401, gin.H{"error": "用户名已存在,请更换用户名..."})
			log.Print("User already exists")
			return
		}
		password, err := utils.HashPassword(user.Password)
		user.Password = password
		if err != nil {
			c.JSON(402, gin.H{"error": "注册失败,请重新再试..."})
		}
		//验证码机制
		code := utils.GenerateCode()
		err = utils.SentEmail(user.Email, "知序验证码", "您的验证码是："+code+"\n该验证码5分钟内有效,请尽快使用。")
		if err != nil {
			c.JSON(403, gin.H{"error": "验证码发送失败,请重新再试..."})
			utils.LogError("验证码发送失败", logrus.Fields{"user_email": user.Email})
			return
		}
		repository.SaveEmailCodeToDB(code, user.Email)
		user.Exist = false
		err = repository.UpdatePasswordByEmail(user.Email, user.Password)
		if err != nil {
			c.JSON(405, gin.H{"error": "请重新再试..."})
			utils.LogError("数据库添加用户失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户注册成功", logrus.Fields{"user_email": user.Email})
		c.JSON(http.StatusOK, gin.H{"message": "修改密码成功!"})
	}
}
