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

// 发表帖子评论
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
		c.JSON(200, gin.H{"success": true})
	}
}

// 删除帖子评论
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

// 点赞flag
func LikeFlag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FlagID uint `json:"flag_id"`
			Like   int  `json:"like"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.UpdateFlagLikes(req.FlagID, req.Like)
		if err != nil {
			c.JSON(500, gin.H{"error": "点赞flag失败,请重新再试..."})
			utils.LogError("数据库点赞flag失败", nil)
			return
		}
	}
}

// 发表flag评论
func CommentOnFlag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var comment model.FlagComment
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.UpdateFlagComment(comment.FlagID, comment.Content)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add comment"})
			utils.LogError("数据库添加flag评论失败", nil)
			return
		}
	}
}

// 删除flag评论
func DeleteFlagComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FlagCommentID uint `json:"flagcomment_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.DeleteFlagComment(req.FlagCommentID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete comment"})
			utils.LogError("数据库删除flag评论失败", nil)
			return
		}
	}
}

// 帖子点赞更改
func LikePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PostID uint `json:"post_id"`
			Like   int  `json:"like"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.UpdatePostLikes(req.PostID, req.Like)
		if err != nil {
			c.JSON(500, gin.H{"error": "点赞帖子失败,请重新再试..."})
			utils.LogError("数据库点赞帖子失败", nil)
			return
		}
	}
}

// 获取flag点赞
func GetFlagLikes() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FlagID uint `json:"flag_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		like, err := repository.GetFlagLikes(req.FlagID)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取flag点赞失败,请重新再试..."})
			utils.LogError("数据库获取flag点赞失败", nil)
			return
		}
		c.JSON(200, gin.H{"like": like})
	}
}

// 获取post点赞
func GetPostLikes() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PostID uint `json:"post_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		like, err := repository.GetPostLikes(req.PostID)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取post点赞失败,请重新再试..."})
			utils.LogError("数据库获取post点赞失败", nil)
			return
		}
		c.JSON(200, gin.H{"like": like})
	}
}
