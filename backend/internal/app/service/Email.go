package service

import (
	"log"
	"net/http"
	"time"

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

		// 更新用户验证状态
		repository.UpdateUserExistStatus(req.Email)

		// 获取用户信息并生成 token
		user, err := repository.GetUserByEmail(req.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取用户信息失败"})
			utils.LogError("验证邮箱后获取用户信息失败", logrus.Fields{"user_email": req.Email})
			return
		}

		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "生成 Token 失败"})
			utils.LogError("验证邮箱后生成token失败", logrus.Fields{"user_email": req.Email})
			return
		}

		utils.LogInfo("邮箱验证成功", logrus.Fields{"user_email": req.Email, "user_id": user.ID})
		utils.SentEmail(req.Email, "邮箱验证成功", "恭喜您成功验证账户")

		c.JSON(200, gin.H{
			"success": true,
			"token":   token,
			"user_id": user.ID,
			"name":    user.Name,
			"email":   user.Email,
		})
	}
}

// 验证码登录（验证邮箱验证码并返回token）
func LoginWithOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			utils.LogError("绑定验证码登录请求参数错误", nil)
			return
		}

		// 验证验证码
		emailCode, err := repository.GetEmailCodeByEmail(req.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取验证码失败,请重新再试..."})
			utils.LogError("获取邮箱验证码失败", nil)
			return
		}
		if emailCode.Code != req.Code {
			c.JSON(400, gin.H{"error": "验证码错误"})
			utils.LogError("邮箱验证码错误", nil)
			return
		}
		if emailCode.Expires.Before(time.Now()) {
			c.JSON(400, gin.H{"error": "验证码已过期"})
			utils.LogError("邮箱验证码已过期", nil)
			return
		}

		// 验证码正确，查找用户
		user, err := repository.GetUserByEmail(req.Email)
		if err != nil || user.ID == 0 {
			c.JSON(404, gin.H{"error": "用户不存在，请先注册"})
			utils.LogError("用户不存在", logrus.Fields{"user_email": req.Email})
			return
		}

		// 生成JWT token
		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "生成token失败,请重新再试..."})
			utils.LogError("生成token失败", logrus.Fields{})
			return
		}

		// 更新用户邮箱验证状态
		repository.UpdateUserExistStatus(req.Email)

		utils.LogInfo("验证码登录成功", logrus.Fields{"user_email": req.Email})
		c.JSON(http.StatusOK, gin.H{
			"token":   token,
			"user_id": user.ID,
			"name":    user.Name,
			"email":   user.Email,
		})
	}
}

func ForgetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData struct {
			Email       string `json:"email"`
			Code        string `json:"code"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(400, gin.H{"error": "请求参数错误,请重新再试..."})
			log.Print("Binding error")
			return
		}

		// 验证密码长度
		if len(requestData.NewPassword) < 6 {
			c.JSON(400, gin.H{"error": "密码长度至少6位"})
			return
		}

		// 验证用户是否存在
		user_exist, _ := repository.GetUserByEmail(requestData.Email)
		if user_exist.ID == 0 {
			c.JSON(404, gin.H{"error": "用户不存在"})
			log.Print("User not found")
			return
		}

		// 验证验证码
		email, err := repository.GetEmailCodeByEmail(requestData.Email)
		if err != nil {
			c.JSON(400, gin.H{"error": "验证码错误或已过期"})
			utils.LogError("获取验证码失败", logrus.Fields{"user_email": requestData.Email})
			return
		}
		if email.Code != requestData.Code {
			c.JSON(400, gin.H{"error": "验证码错误"})
			utils.LogError("验证码错误", logrus.Fields{"user_email": requestData.Email})
			return
		}
		if email.Expires.Before(time.Now()) {
			c.JSON(400, gin.H{"error": "验证码已过期"})
			utils.LogError("验证码已过期", logrus.Fields{"user_email": requestData.Email})
			return
		}

		// 加密新密码
		hashedPassword, err := utils.HashPassword(requestData.NewPassword)
		if err != nil {
			c.JSON(500, gin.H{"error": "密码加密失败,请重新再试..."})
			utils.LogError("密码加密失败", logrus.Fields{})
			return
		}

		// 更新密码
		err = repository.UpdatePasswordByEmail(requestData.Email, hashedPassword)
		if err != nil {
			c.JSON(500, gin.H{"error": "密码更新失败,请重新再试..."})
			utils.LogError("数据库更新密码失败", logrus.Fields{})
			return
		}

		utils.LogInfo("用户密码重置成功", logrus.Fields{"user_email": requestData.Email})
		c.JSON(http.StatusOK, gin.H{"message": "密码重置成功!"})
	}
}

// 发送邮箱验证码
func SendEmailCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "无效的请求参数"})
			utils.LogError("绑定发送验证码请求参数错误", nil)
			return
		}

		// 生成验证码
		code := utils.GenerateCode()

		// 发送邮件
		err := utils.SentEmail(req.Email, "知序验证码", "您的验证码是："+code+"\n该验证码5分钟内有效,请尽快使用。")
		if err != nil {
			c.JSON(500, gin.H{"error": "验证码发送失败,请重新再试..."})
			utils.LogError("验证码发送失败", logrus.Fields{"user_email": req.Email, "error": err.Error()})
			return
		}

		// 保存验证码到数据库
		repository.SaveEmailCodeToDB(code, req.Email)
		utils.LogInfo("验证码发送成功", logrus.Fields{"user_email": req.Email})
		c.JSON(http.StatusOK, gin.H{"message": "验证码已发送!"})
	}
}
