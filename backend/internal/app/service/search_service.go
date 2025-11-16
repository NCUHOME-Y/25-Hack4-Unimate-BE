package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
)

// 更据用户名搜索用户
func SearchUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Keyword string `json:"username"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "错误绑定"})
			return
		}
		users, err := repository.SearchUsers(req.Keyword)
		if err != nil {
			c.JSON(500, gin.H{"error": "搜索用户失败,请重新再试..."})
			utils.LogError("搜索用户失败", nil)
			return
		}
		repository.AddTrackPointToDB(0, "搜索用户")
		c.JSON(200, gin.H{"message": "搜索用户成功", "users": users})
	}
}

// 根据帖子关键字找到对应的帖子
func SearchPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Keyword string `json:"keyword"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "错误绑定"})
			return
		}
		posts, err := repository.SearchPosts(req.Keyword)
		if err != nil {
			c.JSON(500, gin.H{"error": "搜索帖子失败,请重新再试..."})
			utils.LogError("搜索帖子失败", nil)
			return
		}
		repository.AddTrackPointToDB(0, "搜索帖子")
		c.JSON(200, gin.H{"post": posts})
	}
}
