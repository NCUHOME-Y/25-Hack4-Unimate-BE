package service

import (
	"Heckweek/internal/app/model"
	"Heckweek/internal/app/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	repository.DBconnect() //连接数据库
}

var User model.User

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
			return
		}
		user_exist, _ := repository.GetUserByEmail(user_new.Email)
		if user_exist.ID != 0 {
			c.JSON(http.StatusOK, gin.H{"error": "用户名已存在,请更换用户名..."})
			return
		}
		user := model.User{
			Name:     user_new.Name,
			Email:    user_new.Email,
			Password: user_new.Password,
		}
		err := repository.AddUserToDB(user) //添加进数据库
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "注册失败,请重新再试..."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "注册成功!"})
	}
}

// 用户登录
func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_login model.User
		if err := c.ShouldBindJSON(&user_login); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "登录失败,请重新再试..."})
			return
		}
		user_exist, _ := repository.GetUserByEmail(user_login.Email)
		if user_exist.Password != user_login.Password || user_exist.ID == 0 {
			c.JSON(http.StatusOK, gin.H{"error": "用户名或密码错误,请重新再试..."})
			return
		}
		User = user_exist
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

		flags, err := repository.GetFlagsByUserID(User.ID)
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
		flag_current, err := repository.GetFlagsByUserID(User.ID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "获取flag失败,请重新再试..."})
			return
		}
		flag_current = append(flag_current, flag_model)
		err = repository.AddFlagToDB(User, flag_current)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "添加flag失败,请重新再试..."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "添加flag成功!"})
	}
}
