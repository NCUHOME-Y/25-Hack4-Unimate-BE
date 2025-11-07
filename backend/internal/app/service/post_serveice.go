package service

import (
	"Heckweek/internal/app/model"
	"Heckweek/internal/app/repository"

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
			return
		}
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
			return
		}
		c.JSON(200, gin.H{"message": "Post deleted successfully"})
	}
}

// 通过用户name获取他人帖子列表
func GetPostsByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		posts, err := repository.GetPostsByUserName(name)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve posts"})
			return
		}
		c.JSON(200, gin.H{"posts": posts})
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
			return
		}
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
			return
		}
		c.JSON(200, gin.H{"message": "Comment deleted successfully"})
	}
}
