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
		{UserID: user.ID, Name: "首次完成", Description: "第一次设置flag", HadDone: false},
		{UserID: user.ID, Name: "7天连卡", Description: "连续打卡7天", HadDone: false},
		{UserID: user.ID, Name: "任务大师", Description: "完成50个flag", HadDone: false},
		{UserID: user.ID, Name: "目标达成", Description: "积分超过1000", HadDone: false},
		{UserID: user.ID, Name: "学习之星", Description: "累计学习时间超过1000分钟", HadDone: false},
		{UserID: user.ID, Name: "坚持不懈", Description: "累计打卡30天", HadDone: false},
		{UserID: user.ID, Name: "效率达人", Description: "单日完成5个flag", HadDone: false},
		{UserID: user.ID, Name: "专注大师", Description: "单日学习时长超过4小时", HadDone: false},
		{UserID: user.ID, Name: "早起鸟", Description: "早上6点前打卡5次", HadDone: false},
		{UserID: user.ID, Name: "夜猫子", Description: "晚上10点后打卡5次", HadDone: false},
		{UserID: user.ID, Name: "完美主义", Description: "连续10次满分完成flag", HadDone: false},
		{UserID: user.ID, Name: "全能选手", Description: "完成5种不同标签的flag", HadDone: false},
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

		// 转换为前端期望的格式
		type AchievementResponse struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			IsUnlocked  bool   `json:"isUnlocked"`
		}

		result := make([]AchievementResponse, len(achievements))
		for i, a := range achievements {
			result[i] = AchievementResponse{
				ID:          a.ID,
				Name:        a.Name,
				Description: a.Description,
				IsUnlocked:  a.HadDone,
			}
		}

		utils.LogInfo("获取用户成就成功", nil)
		c.JSON(200, gin.H{"message": "获取用户成就成功", "achievements": result})
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
