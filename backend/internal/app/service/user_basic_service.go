package service

import (
	"log"
	"net/http"
	"strings"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 token - 支持 Authorization 头和 URL 参数（用于 WebSocket）
		var token string
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader != "" {
			// 从 Authorization 头获取
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
				log.Printf("[JWT] 从 Authorization 头获取 token")
			} else {
				log.Printf("[JWT] Authorization 格式错误: %s", authHeader)
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请求头中 Authorization 格式有误",
				})
				c.Abort()
				return
			}
		} else {
			// 从 URL 参数获取（用于 WebSocket 连接）
			token = c.Query("token")
			if token == "" {
				log.Printf("[JWT] 未找到 token - Authorization 头为空,URL 参数也为空")
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请求头中 Authorization 为空且 URL 中无 token 参数",
				})
				c.Abort()
				return
			}
			log.Printf("[JWT] 从 URL 参数获取 token: %s...", token[:min(10, len(token))])
		}

		// 解析 token
		claims, err := utils.ParseToken(token)
		if err != nil {
			log.Printf("[JWT] Token 解析失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "无效的 Token",
			})
			c.Abort()
			return
		}

		log.Printf("[JWT] Token 验证成功 - 用户ID: %d, 用户名: %s", claims.UserID, claims.Username)

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("token", token)

		c.Next()
	}
}

func getCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// 类型断言
	id, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return id, true
}

// 用户注册
func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_new struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&user_new); err != nil {
			c.JSON(400, gin.H{"error": "注册失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		user_exist, _ := repository.GetUserByEmail(user_new.Email)
		if user_exist.ID != 0 {
			c.JSON(401, gin.H{"error": "用户名已存在,请更换用户名..."})
			log.Print("User already exists")
			return
		}
		new_password, err := utils.HashPassword(user_new.Password)
		if err != nil {
			c.JSON(402, gin.H{"error": "注册失败,请重新再试..."})
		}
		user := model.User{
			Phone:    user_new.Phone,
			Name:     user_new.Name,
			Email:    user_new.Email,
			Password: new_password,
		}
		user = InitAchievementTable(user)
		if err := repository.AddUserToDB(user); err != nil {
			c.JSON(403, gin.H{"error": "注册失败,请重新再试..."})
			utils.LogError("数据库添加用户失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户注册成功", logrus.Fields{"user_email": user_new.Email})
		c.JSON(http.StatusOK, gin.H{"message": "注册成功!"})
	}
}

// 用户登录
func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_login struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&user_login); err != nil {
			c.JSON(400, gin.H{"error": "登录失败,请重新再试..."})
			return
		}
		user, _ := repository.GetUserByEmail(user_login.Email)
		if !utils.CheckPasswordHash(user_login.Password, user.Password) || user.ID == 0 {
			c.JSON(401, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "生成 Token 失败",
			})
			utils.LogError("生成token失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户登录成功", logrus.Fields{"user_id": user.ID, "user_email": user.Email})
		c.JSON(http.StatusOK, gin.H{"message": "登录成功!",
			"user_id": user.ID,
			"name":    user.Name,
			"email":   user.Email,
			"token":   token})
	}
}

// 更新用户密码
func UpdateUserPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Password    string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		id, _ := getCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		new_token, _ := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(401, gin.H{"error": "请求失败,请重新再试..."})
			utils.LogError("请求绑定失败", logrus.Fields{})
			return
		}
		if !utils.CheckPasswordHash(req.Password, user.Password) {
			c.JSON(400, gin.H{"error": "原密码错误,请重新再试..."})
			return
		}
		req.NewPassword, _ = utils.HashPassword(req.NewPassword)
		err := repository.UpdatePassword(user.ID, req.NewPassword)
		if err != nil {
			c.JSON(500, gin.H{"message": "密码更新失败，请重新再试!"})
			utils.LogError("数据库更新用户数据失败", logrus.Fields{})
			return
		}

		utils.LogInfo("用户密码更新成功", logrus.Fields{"user_id": id})
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"new_token": new_token,
		})
	}
}

// 更新用户名
func UpdateUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NewName string `json:"new_name"`
		}
		id, _ := getCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(501, gin.H{"error": "请求失败,请重新再试..."})
			return
		}
		user, _ := repository.GetUserByID(id)
		if req.NewName == user.Name {
			c.JSON(400, gin.H{"error": "新用户名与原用户名相同,请重新再试..."})
			return
		}
		if req.NewName == "" {
			c.JSON(500, gin.H{"error": "用户名不能为空,请重新再试..."})
			return
		}
		err := repository.UpdateUserName(id, req.NewName)
		if err != nil {
			c.JSON(401, gin.H{"message": "用户名更新失败，请重新再试!"})
			utils.LogError("数据库更新用户数据失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户用户名更新成功", logrus.Fields{"user_id": id, "new_name": req.NewName})
		c.JSON(http.StatusOK, gin.H{
			"success": true})
	}
}

// 更新用户状态
func UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Status string `json:"status"`
		}
		id, _ := getCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新状态失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		err := repository.UpdateUserStatus(id, req.Status)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新状态失败,请重新再试..."})
			utils.LogError("数据库更新用户数据失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户状态更新成功", logrus.Fields{"user_id": id, "new_status": req.Status})
		c.JSON(200, gin.H{
			"message": "状态更新成功",
			"状态":      req.Status})
	}
}

// 获取用户信息
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		user, err := repository.GetUserByID(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取用户状态失败,请重新再试..."})
			utils.LogError("数据库获取用户数据失败", logrus.Fields{})
			return
		}
		utils.LogInfo("获取用户信息成功", logrus.Fields{"user_id": id})
		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

// 打卡
func DoDaKa() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		err := repository.DakaNumberToDB(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "打卡失败,请重新再试..."})
			utils.LogError("数据库更新用户打卡数据失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户打卡成功", logrus.Fields{"user_id": id})
		c.JSON(http.StatusOK, gin.H{"message": "打卡成功!"})
	}
}
