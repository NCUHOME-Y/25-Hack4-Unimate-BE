package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 发布帖子
func PostUserPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		var req struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			FlagID  *uint  `json:"flag_id"` // 关联的Flag ID（可选）
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		post := model.Post{
			Title:   req.Title,
			Content: req.Content,
			FlagID:  req.FlagID,
		}

		err := repository.AddPostToDB(id, post)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add post"})
			utils.LogError("数据库添加帖子失败", nil)
			return
		}
		// 获取刚创建的帖子（包含ID和用户信息）
		posts, _ := repository.GetAllPosts()
		var createdPost model.Post
		if len(posts) > 0 {
			createdPost = posts[0] // 最新的帖子
		}
		utils.LogInfo("用户发布帖子成功", nil)
		c.JSON(200, gin.H{
			"success": true,
			"post":    createdPost,
			"message": "帖子发布成功",
		})
	}
}

// 删除帖子
func DeleteUserPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "未授权"})
			return
		}

		var req struct {
			PostID uint `json:"post_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		// 验证帖子所有权
		post, err := repository.GetPostByID(req.PostID)
		if err != nil {
			c.JSON(404, gin.H{"error": "帖子不存在"})
			return
		}
		if post.UserID != userID {
			c.JSON(403, gin.H{"error": "无权删除此帖子"})
			return
		}

		err = repository.DeletePostFromDB(req.PostID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete post"})
			utils.LogError("数据库删除帖子失败", nil)
			return
		}
		utils.LogInfo("用户删除帖子成功", logrus.Fields{"post_id": req.PostID, "user_id": userID})
		c.JSON(200, gin.H{"success": true})
	}
}

// 发表帖子评论
func CommentOnPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "未授权"})
			return
		}

		var req struct {
			PostID  uint   `json:"postId"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		utils.LogInfo("开始添加评论", logrus.Fields{
			"post_id": req.PostID,
			"user_id": userID,
			"content": req.Content,
		})

		comment := model.PostComment{
			PostID:  req.PostID,
			UserID:  userID,
			Content: req.Content,
		}

		err := repository.AddPostCommentToDB(req.PostID, comment)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add comment"})
			utils.LogError("数据库添加评论失败", logrus.Fields{
				"post_id": req.PostID,
				"error":   err.Error(),
			})
			return
		}

		// 重新查询评论以获取完整的用户信息
		savedComment, err := repository.GetCommentByID(comment.ID)
		if err != nil {
			utils.LogError("查询评论失败", logrus.Fields{"comment_id": comment.ID})
		}

		utils.LogInfo("用户发表评论成功", logrus.Fields{
			"post_id":    req.PostID,
			"comment_id": comment.ID,
		})

		c.JSON(200, gin.H{
			"success":    true,
			"id":         savedComment.ID,
			"userId":     savedComment.UserID,
			"userName":   savedComment.UserName,
			"userAvatar": savedComment.UserAvatar,
			"content":    savedComment.Content,
			"createdAt":  savedComment.CreatedAt.Format("15:04"),
		})
	}
}

// 删除帖子评论
func DeleteUserPostComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "未授权"})
			return
		}

		var req struct {
			CommentID uint `json:"comment_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		// 验证评论所有权
		comment, err := repository.GetCommentByID(req.CommentID)
		if err != nil {
			c.JSON(404, gin.H{"error": "评论不存在"})
			return
		}
		if comment.UserID != userID {
			c.JSON(403, gin.H{"error": "无权删除此评论"})
			return
		}

		err = repository.DeletePostCommentFromDB(req.CommentID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete comment"})
			utils.LogError("数据库删除评论失败", nil)
			return
		}
		utils.LogInfo("用户删除评论成功", logrus.Fields{"comment_id": req.CommentID, "user_id": userID})
		c.JSON(200, gin.H{"success": true, "message": "Comment deleted successfully"})
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
		c.JSON(200, gin.H{
			"success": true,
			"posts":   posts,
			"total":   len(posts),
		})
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
		c.JSON(200, gin.H{
			"success": true,
			"flags":   flags,
			"total":   len(flags),
		})
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
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		// 获取当前用户ID（用于记录点赞关系）
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "未授权"})
			return
		}

		utils.LogInfo("开始处理帖子点赞", logrus.Fields{
			"post_id": req.PostID,
			"user_id": userID,
		})

		// 切换点赞状态（如果已点赞则取消，未点赞则点赞）
		newLikeCount, err := repository.TogglePostLike(req.PostID, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": "点赞帖子失败,请重新再试..."})
			utils.LogError("数据库点赞帖子失败", logrus.Fields{
				"post_id": req.PostID,
				"error":   err.Error(),
			})
			return
		}

		utils.LogInfo("帖子点赞成功", logrus.Fields{
			"post_id":   req.PostID,
			"new_likes": newLikeCount,
		})

		c.JSON(200, gin.H{
			"success": true,
			"likes":   newLikeCount,
		})
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

// 获取当前用户点过赞的帖子ID
func GetUserLikedPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "未授权"})
			return
		}
		ids, err := repository.GetLikedPostIDsByUser(userID)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取已点赞帖子失败"})
			utils.LogError("获取已点赞帖子失败", nil)
			return
		}
		c.JSON(200, gin.H{"liked_post_ids": ids})
	}
}
