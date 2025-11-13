package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// 记录学习时长
func RecordLearnTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		cron := cron.New()
		cron.AddFunc("@daily", func() {
			err := repository.AddNewLearnTimeToDB(id)
			if err != nil {
				utils.LogError("添加新的学习时间记录失败", nil)
				return
			}
			utils.LogInfo("添加新的学习时间记录成功", nil)
		})
		cron.Start()
		var req struct {
			Duration int `json:"duration"` // 学习时长，单位为分钟
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		err := repository.UpdateLearnTimeDuration(id, req.Duration)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update learn time"})
			utils.LogError("更新学习时间记录失败", nil)
			return
		}
	}
}

// 获取一个月的学习时长记录
func GetLearnTimeRecords() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		learnTimes, err := repository.GetRecentLearnTime(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取学习时长记录失败,请重新再试..."})
			utils.LogError("获取学习时长记录失败", logrus.Fields{"user_id": id})
			return
		}
		utils.LogInfo("获取学习时长记录成功", logrus.Fields{"user_id": id})
		c.JSON(200, gin.H{
			"learn_times": learnTimes,
		})
	}
}

// 获取打卡总数
func GetUserDakaTotal() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		c.JSON(200, gin.H{
			"daka_total": user.Daka,
		})
	}
}

// 获取月打卡数
func GetUserMonthDaka() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		dakaNumber, _ := repository.GetRecentDakaNumber(id)
		c.JSON(200, gin.H{
			"month_daka": dakaNumber.MonthDaka,
		})
	}
}
