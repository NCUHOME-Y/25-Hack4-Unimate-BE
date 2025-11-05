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

// 用户注册
func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_new model.User
		if err := c.ShouldBindJSON(&user_new); err != nil {
			c.JSON(http.StatusOK, gin.H{"错误": "注册失败,请重新再试..."})
			return
		}
		user_exist, _ := repository.GetUserByName(user_new.Name)
		if user_exist.ID != 0 {
			c.JSON(http.StatusOK, gin.H{"error": "用户名已存在,请更换用户名..."})
			return
		}
		err := repository.AddUserToDB(user_new) //添加进数据库
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "注册失败,请重新再试..."})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "注册成功!"})
			return
		}
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
		user_exist, _ := repository.GetUserByName(user_login.Name)
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
