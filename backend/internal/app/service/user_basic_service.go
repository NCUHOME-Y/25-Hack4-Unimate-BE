package service

import (
	"Heckweek/internal/app/model"
	"Heckweek/internal/app/repository"
	utils "Heckweek/util"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 token
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "请求头中 Authorization 为空",
			})
			c.Abort()
			return
		}

		// 检查 token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "请求头中 Authorization 格式有误",
			})
			c.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "无效的 Token",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

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
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&user_new); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "注册失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		user_exist, _ := repository.GetUserByEmail(user_new.Email)
		if user_exist.ID != 0 {
			c.JSON(http.StatusOK, gin.H{"error": "用户名已存在,请更换用户名..."})
			log.Print("User already exists")
			return
		}
		user := model.User{
			Name:     user_new.Name,
			Email:    user_new.Email,
			Password: user_new.Password,
		}
		if err := repository.AddUserToDB(user); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "注册失败,请重新再试..."})
			log.Print("Add user to DB error")
			return
		}
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
			c.JSON(http.StatusOK, gin.H{"error": "登录失败,请重新再试..."})
			return
		}
		user_exist, _ := repository.GetUserByEmail(user_login.Email)
		if user_exist.Password != user_login.Password || user_exist.ID == 0 {
			c.JSON(http.StatusOK, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "登录成功!"})
	}
}

// 更新用户密码
func UpdateUserPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID          uint   `json:"id"`
			Password    string `json:"password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "请求失败,请重新再试..."})
			return
		}
		user, _ := repository.GetUserByID(req.ID)
		if req.Password != user.Password {
			c.JSON(http.StatusOK, gin.H{"error": "原密码错误,请重新再试..."})
			return
		}
		repository.UpdatePassword(req.ID, req.NewPassword)
		c.JSON(http.StatusOK, gin.H{"message": "密码更新成功!"})
	}
}

func UpdateUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID      uint   `json:"id"`
			NewName string `json:"new_name"`
		}
		user, _ := repository.GetUserByID(req.ID)
		if user.Name == req.NewName {
			c.JSON(http.StatusOK, gin.H{"error": "新用户名与原用户名相同,请重新再试..."})
			return
		}
		if req.NewName == "" {
			c.JSON(http.StatusOK, gin.H{"error": "用户名不能为空,请重新再试..."})
			return
		}
		err := repository.UpdateUserName(req.ID, req.NewName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "用户名更新失败，请重新再试!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "用户名更新成功!"})
	}
}

// 获取用户flag
func GetUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		flags, err := repository.GetFlagsByUserID(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "获取flag失败,请重新再试..."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}

func PostUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var flag struct {
			Flag        string `json:"flag"`
			PlanContent string `json:"plan_content"`
			IsHiden     bool   `json:"is_hiden"`
		}
		if err := c.ShouldBindJSON(&flag); err != nil {
			c.JSON(http.StatusOK, gin.H{"错误": "添加flag失败,请重新再试..."})
			return
		}
		flag_model := model.Flag{
			Flag:        flag.Flag,
			PlanContent: flag.PlanContent,
			IsHiden:     flag.IsHiden,
		}
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		flag_current, err := repository.GetFlagsByUserID(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "获取flag失败,请重新再试..."})
			return
		}
		flag_current = append(flag_current, flag_model)
		err = repository.AddFlagToDB(id, flag_current)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "添加flag失败,请重新再试..."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "添加flag成功!"})
	}
}
