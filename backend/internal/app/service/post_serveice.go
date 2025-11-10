package service

import (
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"

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
		c.JSON(200, gin.H{"message": "Post added successfully"})
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
		c.JSON(200, gin.H{"message": "Post deleted successfully"})
	}
}

// 发表评论
func CommentOnPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var comment model.PostComment
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.AddPostCommentToDB(comment.PostID, comment)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add comment"})
			utils.LogError("数据库添加评论失败", nil)
			return
		}
		utils.LogInfo("用户发表评论成功", nil)
		c.JSON(200, gin.H{"message": "Comment added successfully"})
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
