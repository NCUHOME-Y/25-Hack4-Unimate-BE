package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
)

// 发布帖子
func PostUserPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		var post model.Post
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.AddPostToDB(id, post)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add post"})
			utils.LogError("数据库添加帖子失败", nil)
			return
		}
		utils.LogInfo("用户发布帖子成功", nil)
		c.JSON(200, gin.H{"success": true})
	}
}

// 删除帖子
func DeleteUserPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PostID uint `json:"post_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.DeletePostFromDB(req.PostID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete post"})
			utils.LogError("数据库删除帖子失败", nil)
			return
		}
		utils.LogInfo("用户删除帖子成功", nil)
		c.JSON(200, gin.H{"success": true})
	}
}

// 发表评论
func CommentOnPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var comment model.Comment
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.AddPostCommentToDB(comment.CommentID, comment)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add comment"})
			utils.LogError("数据库添加评论失败", nil)
			return
		}
		utils.LogInfo("用户发表评论成功", nil)
		c.JSON(200, gin.H{"success": true})
	}
}

// 删除评论
func DeleteUserPostComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CommentID uint `json:"comment_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.DeletePostCommentFromDB(req.CommentID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete comment"})
			utils.LogError("数据库删除评论失败", nil)

			return
		}
		utils.LogInfo("用户删除评论成功", nil)
		c.JSON(200, gin.H{"message": "Comment deleted successfully"})
	}
}

// 获取所有帖子
func GetAllPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := repository.GetAllPosts()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve posts"})
			utils.LogError("数据库获取帖子失败", nil)
			return
		}
		utils.LogInfo("获取所有帖子成功", nil)
		c.JSON(200, gin.H{"posts": posts})
	}
}

// 获取所有可见的flag
func GetVisibleFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		flags, err := repository.GetVisibleFlags()
		if err != nil {
			c.JSON(500, gin.H{"error": "获取可见flag失败,请重新再试..."})
			utils.LogError("数据库获取可见flag失败", nil)
			return
		}
		utils.LogInfo("获取所有可见flag成功", nil)
		c.JSON(200, gin.H{"flags": flags})
	}
}
