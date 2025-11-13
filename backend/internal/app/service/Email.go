package service

import (
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
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
		c.JSON(200, gin.H{"success": true})
	}
}
