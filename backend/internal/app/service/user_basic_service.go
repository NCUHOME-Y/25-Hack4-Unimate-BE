package service

import (
	"Heckweek/internal/app/model"
	"Heckweek/internal/app/repository"
	utils "Heckweek/util"
	"log"
	"net/http"
	"strings"
	"time"

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
		user, _ := repository.GetUserByEmail(user_login.Email)
		if user.Password != user_login.Password || user.ID == 0 {
			c.JSON(http.StatusOK, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		token, err := utils.GenerateToken(user.ID, user.Name, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "生成 Token 失败",
			})
			log.Print("Generate token error")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "登录成功!",
			"token": token})
	}
}

// 更新用户密码
func UpdateUserPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email       uint   `json:"id"`
			Password    string `json:"password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(401, gin.H{"error": "请求失败,请重新再试..."})
			return
		}
		user, _ := repository.GetUserByID(req.Email)
		if req.Password != user.Password {
			c.JSON(400, gin.H{"error": "原密码错误,请重新再试..."})
			return
		}
		err := repository.UpdatePassword(user.ID, req.NewPassword)
		if err != nil {
			c.JSON(500, gin.H{"message": "密码更新失败，请重新再试!"})
			log.Printf(" Password update error: %v", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "密码更新成功!"})
	}
}

func UpdateUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID      uint   `json:"id"`
			NewName string `json:"new_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(501, gin.H{"error": "请求失败,请重新再试..."})
			return
		}
		user, _ := repository.GetUserByID(req.ID)
		if req.NewName == user.Name {
			c.JSON(400, gin.H{"error": "新用户名与原用户名相同,请重新再试..."})
			return
		}
		if req.NewName == "" {
			c.JSON(500, gin.H{"error": "用户名不能为空,请重新再试..."})
			return
		}
		err := repository.UpdateUserName(req.ID, req.NewName)
		if err != nil {
			c.JSON(401, gin.H{"message": "用户名更新失败，请重新再试!"})
			log.Printf(" Username update error: %v", err)
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
			log.Print("Get flags error")
			return
		}
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}

// 添加用户flag
func PostUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var flag struct {
			Flag           string `json:"flag"`
			PlanContent    string `json:"plan_content"`
			IsHiden        bool   `json:"is_hiden"`
			PlanDoneNumber int    `json:"plan_done_number"`
			DeadTime       string `json:"deadtime"`
		}
		if err := c.ShouldBindJSON(&flag); err != nil {
			c.JSON(500, gin.H{"err": "添加flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		t, _ := time.Parse(flag.DeadTime, "2006-01-02 15:04:05")
		flag_model := model.Flag{
			Flag:           flag.Flag,
			PlanContent:    flag.PlanContent,
			IsHiden:        flag.IsHiden,
			PlanDoneNumber: flag.PlanDoneNumber,
			DeadTime:       t,
		}
		id, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(402, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		err := repository.AddFlagToDB(id, flag_model)
		if err != nil {
			c.JSON(400, gin.H{"error": "添加flag失败,请重新再试..."})
			log.Print("Add flag to DB error")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "添加flag成功!"})
	}
}

// 完成用户flag
func DoneUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID         uint `json:"id"`
			DoneNumber int  `json:"done_number"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		err := repository.UpdateFlagDoneNumber(req.ID, req.DoneNumber)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			return
		}
		c.JSON(200, gin.H{"message": "打卡成功"})
	}
}
