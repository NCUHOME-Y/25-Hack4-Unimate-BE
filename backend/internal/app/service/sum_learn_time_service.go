package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 记录学习时长
func RecordLearnTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)

		var req struct {
			Duration int `json:"duration"` // 学习时长，单位为秒
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数格式错误"})
			utils.LogError("解析学习时长参数失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 更新learn_times表（每日记录）
		err := repository.UpdateLearnTimeDuration(id, req.Duration)
		if err != nil {
			c.JSON(500, gin.H{"error": "记录学习时长失败"})
			utils.LogError("更新学习时间记录失败", logrus.Fields{"user_id": id, "duration": req.Duration, "error": err.Error()})
			return
		}

		// 更新用户的month_learntime（累计本月学习时长）
		user, err := repository.GetUserByID(id)
		if err == nil {
			newMonthTime := user.MonthLearntime + req.Duration
			err = repository.DB.Model(&user).Update("month_learntime", newMonthTime).Error
			if err != nil {
				utils.LogError("更新用户月学习时长失败", logrus.Fields{"user_id": id, "error": err.Error()})
			} else {
				utils.LogInfo("更新用户月学习时长成功", logrus.Fields{"user_id": id, "duration": req.Duration, "total": newMonthTime})
			}
		}

		utils.LogInfo("记录学习时长成功", logrus.Fields{"user_id": id, "duration": req.Duration})
		c.JSON(200, gin.H{"success": true, "message": "学习时长已记录", "duration": req.Duration})
	}
}

// 获取最近一个月的学习时长记录
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

// 获取最近7天的数据
func GetLearnTimeLast7Days() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		learnTimes, err := repository.GetSevenDaysLearnTime(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取最近7天学习时长记录失败,请重新再试..."})
			utils.LogError("获取最近7天学习时长记录失败", logrus.Fields{"user_id": id})
			return
		}
		utils.LogInfo("获取最近7天学习时长记录成功", logrus.Fields{"user_id": id})
		c.JSON(200, gin.H{
			"learn_times": learnTimes,
		})
	}
}

// 获取最近180的数据
func GetLearnTimeLast180Days() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		learnTimes, err := repository.GetRecent180LearnTime(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取最近180天学习时长记录失败,请重新再试..."})
			utils.LogError("获取最近180天学习时长记录失败", logrus.Fields{"user_id": id})
		}
		utils.LogInfo("获取最近180天学习时长记录成功", logrus.Fields{"user_id": id})
		c.JSON(200, gin.H{
			"learn_times": learnTimes,
		})
	}
}

// 获取当前月份的学习时长记录
func GetCurrentMonthLearnTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		learnTimes, err := repository.GetCurrentMonthLearnTime(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取当前月份学习时长记录失败,请重新再试..."})
			utils.LogError("获取当前月份学习时长记录失败", logrus.Fields{"user_id": id})
			return
		}
		utils.LogInfo("获取当前月份学习时长记录成功", logrus.Fields{"user_id": id})
		c.JSON(200, gin.H{
			"learn_times": learnTimes,
		})
	}
}

// 获取最近6个月的数据
func GetRecent6MonthsLearnTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		learnTimes, err := repository.GetRecent6MonthsLearnTime(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取最近6个月学习时长记录失败,请重新再试..."})
			utils.LogError("获取最近6个月学习时长记录失败", logrus.Fields{"user_id": id})
			return
		}
		utils.LogInfo("获取最近6个月学习时长记录成功", logrus.Fields{"user_id": id})
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

// 月学习时长
func GetLearnTimeRecordsMonth() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		user, _ := repository.GetUserByID(id)
		c.JSON(200, gin.H{
			"month_learntime": user.MonthLearntime,
		})
	}
}

// 完成flag的标签数种类
func GetLabelByUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		labal, err := repository.GetLabelByUserID(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取用户标签失败,请重新再试..."})
			utils.LogError("获取用户标签失败", logrus.Fields{"user_id": id})
			return
		}
		utils.LogInfo("获取用户标签成功", logrus.Fields{"user_id": id})
		c.JSON(200, gin.H{
			"label": labal,
		})
	}
}

// 🔧 新增：获取今日学习时长
func GetTodayLearnTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		learnTime, err := repository.GetTodayLearnTime(id)
		if err != nil {
			c.JSON(200, gin.H{
				"today_learn_time": 0,
			})
			return
		}
		utils.LogInfo("获取今日学习时长成功", logrus.Fields{"user_id": id, "duration": learnTime.Duration})
		c.JSON(200, gin.H{
			"today_learn_time": learnTime.Duration,
		})
	}
}
