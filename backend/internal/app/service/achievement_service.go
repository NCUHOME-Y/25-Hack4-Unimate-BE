package service

import (
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func InitAchievementTable(user model.User) model.User {
	achievements := []model.Achievement{
		{UserID: user.ID, Name: "新手启程", Description: "第一次设置flag", HadDone: false},
		{UserID: user.ID, Name: "坚持不懈", Description: "第一次坚持打卡完成", HadDone: false},
		{UserID: user.ID, Name: "学习达人", Description: "累计打卡50次", HadDone: false},
		{UserID: user.ID, Name: "任务收藏家", Description: "积分超过1000", HadDone: false},
		{UserID: user.ID, Name: "专注家", Description: "累计学习时间超过1000min", HadDone: false},
	}
	user.Achievements = achievements
	utils.LogInfo("初始化用户成就表成功", nil)
	return user
}

// 调取用户成就
func GetUserAchievement() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := getCurrentUserID(c)
		AchievementCheckAll(id)
		achievements, err := repository.GetAchievementsByUserID(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取用户成就失败,请重新再试..."})
			utils.LogError("获取用户成就失败", nil)
			return
		}
		utils.LogInfo("获取用户成就成功", nil)
		c.JSON(200, gin.H{"message": "获取用户成就成功", "data": achievements})
	}
}

// 成就检测合集
func AchievementCheckAll(userID uint) {
	AchievementCheckFirstFlag(userID)
	AchievementCheckFirstKeepFlag(userID)
	AchievementCheckLearn50Days(userID)
	AchievementCheckCompleteFlag100Times(userID)
	AchievementCheckLearn1000Min(userID)
}

// 成就检测：新手启程
func AchievementCheckFirstFlag(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if len(user.Flags) >= 1 {
		err := repository.UpdateAchievementHadDone(userID, "新手启程")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "新手启程"})
			return
		}
	}
}

// 成就检测：坚持不懈
func AchievementCheckFirstKeepFlag(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if user.FlagNumber >= 1 {
		err := repository.UpdateAchievementHadDone(userID, "坚持不懈")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "坚持不懈"})
			return
		}
	}
}

// 成就检测：学习达人
func AchievementCheckLearn50Days(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if user.FlagNumber >= 50 {
		err := repository.UpdateAchievementHadDone(userID, "学习达人")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "学习达人"})
			return
		}
	}
}

// 成就检测：任务收藏家
func AchievementCheckCompleteFlag100Times(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if user.Count >= 1000 {
		err := repository.UpdateAchievementHadDone(userID, "任务收藏家")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "任务收藏家"})
			return
		}
	}
}

// 成就检测：专注家
func AchievementCheckLearn1000Min(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	var totalLearnTime int
	for _, learnTime := range user.LearnTimes {
		totalLearnTime += learnTime.Duration
	}
	if totalLearnTime >= 1000 {
		err := repository.UpdateAchievementHadDone(userID, "专注家")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "专注家"})
			return
		}
	}
}
