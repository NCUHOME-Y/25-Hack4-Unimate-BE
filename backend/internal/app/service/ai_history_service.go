package service

import (
	"encoding/json"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 保存AI历史记录
func SaveAIHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok || id == 0 {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}

		var req struct {
			Background    string `json:"background"`
			Goal          string `json:"goal"`
			Difficulty    string `json:"difficulty"`
			GeneratedPlan string `json:"generated_plan"` // JSON字符串
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数格式错误"})
			utils.LogError("解析AI历史记录参数失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 创建AI历史记录
		aiHistory := model.AIHistory{
			UserID:        id,
			Background:    req.Background,
			Goal:          req.Goal,
			Difficulty:    req.Difficulty,
			GeneratedPlan: req.GeneratedPlan,
			CreatedAt:     time.Now(),
		}

		if err := repository.DB.Create(&aiHistory).Error; err != nil {
			c.JSON(500, gin.H{"error": "保存AI历史记录失败"})
			utils.LogError("保存AI历史记录失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}

		utils.LogInfo("保存AI历史记录成功", logrus.Fields{"user_id": id, "goal": req.Goal})
		c.JSON(200, gin.H{"success": true, "message": "AI历史记录已保存"})
	}
}

// 获取AI历史记录
func GetAIHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok || id == 0 {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}

		var aiHistories []model.AIHistory
		if err := repository.DB.Where("user_id = ?", id).Order("created_at desc").Limit(10).Find(&aiHistories).Error; err != nil {
			c.JSON(500, gin.H{"error": "获取AI历史记录失败"})
			utils.LogError("获取AI历史记录失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}

		// 转换为前端需要的格式
		var result []gin.H
		for _, history := range aiHistories {
			var plan interface{}
			if err := json.Unmarshal([]byte(history.GeneratedPlan), &plan); err != nil {
				// 如果解析失败，使用原始字符串
				plan = history.GeneratedPlan
			}

			result = append(result, gin.H{
				"id":             history.ID,
				"background":     history.Background,
				"goal":           history.Goal,
				"difficulty":     history.Difficulty,
				"generated_plan": plan,
				"created_at":     history.CreatedAt,
			})
		}

		utils.LogInfo("获取AI历史记录成功", logrus.Fields{"user_id": id, "count": len(result)})
		c.JSON(200, gin.H{"ai_histories": result})
	}
}

// 删除AI历史记录
func DeleteAIHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok || id == 0 {
			c.JSON(401, gin.H{"error": "未授权，请先登录"})
			return
		}

		var req struct {
			HistoryID uint `json:"history_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数格式错误"})
			utils.LogError("解析删除AI历史记录参数失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 确保只能删除自己的记录
		if err := repository.DB.Where("id = ? AND user_id = ?", req.HistoryID, id).Delete(&model.AIHistory{}).Error; err != nil {
			c.JSON(500, gin.H{"error": "删除AI历史记录失败"})
			utils.LogError("删除AI历史记录失败", logrus.Fields{"user_id": id, "history_id": req.HistoryID, "error": err.Error()})
			return
		}

		utils.LogInfo("删除AI历史记录成功", logrus.Fields{"user_id": id, "history_id": req.HistoryID})
		c.JSON(200, gin.H{"success": true, "message": "AI历史记录已删除"})
	}
}
