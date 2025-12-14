package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 手动触发一言发布（用于测试）
func TriggerHitokoto() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 确保一言账户存在
		if err := ensureHitokotoUser(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "一言账户初始化失败: " + err.Error(),
			})
			return
		}

		// 发布一言
		if err := postHitokoto(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "发布一言失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "一言发布成功",
		})
	}
}
