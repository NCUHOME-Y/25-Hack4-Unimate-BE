package service

import (
	"fmt"
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
		password, _ := c.Get("user_password")
		user, _ := repository.GetUserByEmail(req.Email)
		user.Password = password.(string)
		err = repository.SaveUserToDB(user)
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
			c.JSON(400, gin.H{"error": "请求参数错误"})
			utils.LogError("验证码登录参数绑定错误", logrus.Fields{"error": err.Error()})
			return
		}

		// 验证邮箱格式
		if !validateEmail(req.Email) {
			c.JSON(400, gin.H{"error": "邮箱格式不正确"})
			return
		}

		// 验证验证码
		emailCode, err := repository.GetEmailCodeByEmail(req.Email)
		if err != nil {
			c.JSON(400, gin.H{"error": "验证码不存在或已过期"})
			utils.LogWarn("获取邮箱验证码失败", logrus.Fields{"email": req.Email})
			return
		}
		if !secureCompareCode(emailCode.Code, req.Code) {
			c.JSON(400, gin.H{"error": "验证码错误"})
			utils.LogWarn("验证码错误", logrus.Fields{"email": req.Email})
			return
		}
		if emailCode.Expires.Before(time.Now()) {
			c.JSON(400, gin.H{"error": "验证码已过期"})
			utils.LogWarn("验证码已过期", logrus.Fields{"email": req.Email})
			return
		}

		// 验证码正确，查找用户
		user, err := repository.GetUserByEmail(req.Email)
		if err != nil || user.ID == 0 {
			c.JSON(400, gin.H{"error": "该邮箱尚未注册，请先注册账号"})
			utils.LogInfo("验证码登录失败-用户不存在", logrus.Fields{"email": req.Email})
			return
		}

		// 删除已使用的验证码
		if err := repository.DeleteEmailCode(req.Email); err != nil {
			utils.LogWarn("删除验证码失败", logrus.Fields{"email": req.Email, "error": err.Error()})
		}

		// 生成JWT token
		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "生成token失败"})
			utils.LogError("生成token失败", logrus.Fields{"user_id": user.ID, "error": err.Error()})
			return
		}

		// 更新用户邮箱验证状态
		repository.UpdateUserExistStatus(req.Email)

		utils.LogInfo("验证码登录成功", logrus.Fields{"email": req.Email, "user_id": user.ID})
		c.JSON(http.StatusOK, gin.H{
			"message":          "登录成功！",
			"token":            token,
			"user_id":          user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"head_show":        user.HeadShow,
			"daka":             user.Daka,
			"flag_number":      user.FlagNumber,
			"count":            user.Count,
			"month_learn_time": user.MonthLearntime,
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
			c.JSON(400, gin.H{"error": "请求参数错误"})
			utils.LogError("密码重置参数绑定错误", logrus.Fields{"error": err.Error()})
			return
		}

		// 验证输入
		if !validateEmail(requestData.Email) {
			c.JSON(400, gin.H{"error": "邮箱格式不正确"})
			return
		}
		if !validatePassword(requestData.NewPassword) {
			c.JSON(400, gin.H{"error": "密码长度需在6-100个字符之间"})
			return
		}

		// 验证用户是否存在
		user_exist, _ := repository.GetUserByEmail(requestData.Email)
		if user_exist.ID == 0 {
			c.JSON(400, gin.H{"error": "该邮箱尚未注册"})
			utils.LogInfo("密码重置失败-用户不存在", logrus.Fields{"email": requestData.Email})
			return
		}

		// 验证验证码
		email, err := repository.GetEmailCodeByEmail(requestData.Email)
		if err != nil {
			c.JSON(400, gin.H{"error": "验证码不存在或已过期"})
			utils.LogWarn("获取验证码失败", logrus.Fields{"email": requestData.Email})
			return
		}
		if !secureCompareCode(email.Code, requestData.Code) {
			c.JSON(400, gin.H{"error": "验证码错误"})
			utils.LogWarn("验证码错误", logrus.Fields{"email": requestData.Email})
			return
		}
		if email.Expires.Before(time.Now()) {
			c.JSON(400, gin.H{"error": "验证码已过期"})
			utils.LogWarn("验证码已过期", logrus.Fields{"email": requestData.Email})
			return
		}

		// 删除已使用的验证码
		if err := repository.DeleteEmailCode(requestData.Email); err != nil {
			utils.LogWarn("删除验证码失败", logrus.Fields{"email": requestData.Email, "error": err.Error()})
		}

		// 加密新密码
		hashedPassword, err := utils.HashPassword(requestData.NewPassword)
		if err != nil {
			c.JSON(500, gin.H{"error": "密码加密失败"})
			utils.LogError("密码加密失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 更新密码
		err = repository.UpdatePasswordByEmail(requestData.Email, hashedPassword)
		if err != nil {
			c.JSON(500, gin.H{"error": "密码更新失败"})
			utils.LogError("数据库更新密码失败", logrus.Fields{"error": err.Error()})
			return
		}

		utils.LogInfo("用户密码重置成功", logrus.Fields{"email": requestData.Email, "user_id": user_exist.ID})
		c.JSON(http.StatusOK, gin.H{"message": "密码重置成功！"})
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

		// 检查发送频率限制（1分钟内只能发送一次）
		canSend, lastSentTime, err := repository.CheckEmailCodeRateLimit(req.Email)
		if err != nil {
			c.JSON(500, gin.H{"error": "检查发送频率失败,请重新再试..."})
			utils.LogError("检查验证码发送频率失败", logrus.Fields{"user_email": req.Email, "error": err.Error()})
			return
		}
		if !canSend {
			// 计算还需要等待多少秒
			waitSeconds := 60 - int(time.Since(lastSentTime).Seconds())
			if waitSeconds < 0 {
				waitSeconds = 0
			}
			c.JSON(429, gin.H{
				"error":        "发送过于频繁,请稍后再试",
				"wait_seconds": waitSeconds,
				"message":      fmt.Sprintf("请等待%d秒后再试", waitSeconds),
			})
			utils.LogInfo("验证码发送被限流", logrus.Fields{"user_email": req.Email, "wait_seconds": waitSeconds})
			return
		}

		// 生成验证码
		code := utils.GenerateCode()

		// 发送邮件
		emailBody := fmt.Sprintf(`尊敬的用户，您好！

您正在进行身份验证，验证码为：

    %s

该验证码将在5分钟内有效，请及时使用。如非本人操作，请忽略此邮件。

请登录知序平台（http://139.199.157.76）查看详情。

——知序平台`, code)
		sendErr := utils.SentEmail(req.Email, "【知序】身份验证码", emailBody)
		if sendErr != nil {
			c.JSON(500, gin.H{"error": "验证码发送失败,请重新再试..."})
			utils.LogError("验证码发送失败", logrus.Fields{"user_email": req.Email, "error": sendErr.Error()})
			return
		}

		// 保存验证码到数据库
		repository.SaveEmailCodeToDB(code, req.Email)
		utils.LogInfo("验证码发送成功", logrus.Fields{"user_email": req.Email})
		c.JSON(http.StatusOK, gin.H{"message": "验证码已发送!"})
	}
}
