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
		{UserID: user.ID, Name: "学习狂人", Description: "累计学习时间超过5000分钟", HadDone: false},
		{UserID: user.ID, Name: "社交达人", Description: "发布10条动态", HadDone: false},
		{UserID: user.ID, Name: "时间管理者", Description: "连续30天完成至少1个flag", HadDone: false},
		{UserID: user.ID, Name: "成就收集者", Description: "解锁10个徽章", HadDone: false},
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

		// 补全缺失的成就
		allAchievementList := []struct{ Name, Description string }{
			{"首次完成", "第一次设置flag"},
			{"7天连卡", "连续打卡7天"},
			{"任务大师", "完成50个flag"},
			{"目标达成", "积分超过1000"},
			{"学习之星", "累计学习时间超过1000分钟"},
			{"坚持不懈", "累计打卡30天"},
			{"效率达人", "单日完成5个flag"},
			{"专注大师", "单日学习时长超过4小时"},
			{"早起鸟", "早上6点前打卡5次"},
			{"夜猫子", "晚上10点后打卡5次"},
			{"完美主义", "连续10次满分完成flag"},
			{"全能选手", "完成5种不同标签的flag"},
			{"学习狂人", "累计学习时间超过5000分钟"},
			{"社交达人", "发布10条动态"},
			{"时间管理者", "连续30天完成至少1个flag"},
			{"成就收集者", "解锁10个徽章"},
		}
		existMap := make(map[string]bool)
		for _, a := range achievements {
			existMap[a.Name] = true
		}
		for _, item := range allAchievementList {
			if !existMap[item.Name] {
				// 插入缺失的成就
				_ = repository.InsertAchievement(id, item.Name, item.Description)
			}
		}
		// 重新查询补全后的成就
		achievements, _ = repository.GetAchievementsByUserID(id)

		// 转换为前端期望的格式
		type AchievementResponse struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			IsUnlocked  bool   `json:"isUnlocked"`
		}

		// 去重：使用map确保成就名称唯一
		uniqueAchievements := make(map[string]model.Achievement)
		for _, a := range achievements {
			// 如果成就名称已存在，保留ID较小的记录（通常是较早创建的）
			if existing, exists := uniqueAchievements[a.Name]; !exists || a.ID < existing.ID {
				uniqueAchievements[a.Name] = a
			}
		}

		// 转换为切片
		result := make([]AchievementResponse, 0, len(uniqueAchievements))
		for _, a := range uniqueAchievements {
			result = append(result, AchievementResponse{
				ID:          a.ID,
				Name:        a.Name,
				Description: a.Description,
				IsUnlocked:  a.HadDone,
			})
		}

		utils.LogInfo("获取用户成就成功", nil)
		c.JSON(200, gin.H{"message": "获取用户成就成功", "achievements": result})
	}
}

// 成就检测合集
func AchievementCheckAll(userID uint) {
	AchievementCheckFirstFlag(userID)            // 首次完成
	AchievementCheckFirstKeepFlag(userID)        // 7天连卡
	AchievementCheckLearn50Days(userID)          // 任务大师
	AchievementCheckCompleteFlag100Times(userID) // 目标达成
	AchievementCheckLearn1000Min(userID)         // 学习之星
	AchievementCheckDaka30Days(userID)           // 坚持不懈
	AchievementCheckDailyFlag5(userID)           // 效率达人
	AchievementCheckDailyLearn4Hours(userID)     // 专注大师
	AchievementCheckEarlyBird(userID)            // 早起鸟
	AchievementCheckNightOwl(userID)             // 夜猫子
	AchievementCheckPerfectStreak(userID)        // 完美主义
	AchievementCheckAllRounder(userID)           // 全能选手
	AchievementCheckLearn5000Min(userID)         // 学习狂人
	AchievementCheckPost10Times(userID)          // 社交达人
	AchievementCheckDailyFlag30Days(userID)      // 时间管理者
	AchievementCheckUnlock10Badges(userID)       // 成就收集者
}

// 成就检测：首次完成
func AchievementCheckFirstFlag(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if len(user.Flags) >= 1 {
		err := repository.UpdateAchievementHadDone(userID, "首次完成")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "首次完成"})
			return
		}
	}
}

// 成就检测：7天连卡
func AchievementCheckFirstKeepFlag(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现连续打卡7天的检测逻辑
	if user.Daka >= 7 {
		err := repository.UpdateAchievementHadDone(userID, "7天连卡")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "7天连卡"})
			return
		}
	}
}

// 成就检测：任务大师
func AchievementCheckLearn50Days(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if user.FlagNumber >= 50 {
		err := repository.UpdateAchievementHadDone(userID, "任务大师")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "任务大师"})
			return
		}
	}
}

// 成就检测：目标达成
func AchievementCheckCompleteFlag100Times(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if user.Count >= 1000 {
		err := repository.UpdateAchievementHadDone(userID, "目标达成")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "目标达成"})
			return
		}
	}
}

// 成就检测：学习之星
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
		err := repository.UpdateAchievementHadDone(userID, "学习之星")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "学习之星"})
			return
		}
	}
}

// 成就检测：坚持不懈（累计打卡30天）
func AchievementCheckDaka30Days(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	if user.Daka >= 30 {
		err := repository.UpdateAchievementHadDone(userID, "坚持不懈")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "坚持不懈"})
			return
		}
	}
}

// 成就检测：效率达人（单日完成5个flag）
func AchievementCheckDailyFlag5(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现单日完成flag数量统计
	// 暂时使用总完成数作为判断条件
	if user.FlagNumber >= 5 {
		err := repository.UpdateAchievementHadDone(userID, "效率达人")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "效率达人"})
			return
		}
	}
}

// 成就检测：专注大师（单日学习时长超过4小时=240分钟）
func AchievementCheckDailyLearn4Hours(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现单日学习时长统计
	// 暂时使用本月学习时长作为判断条件
	if user.MonthLearntime >= 240 {
		err := repository.UpdateAchievementHadDone(userID, "专注大师")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "专注大师"})
			return
		}
	}
}

// 成就检测：早起鸟（早上6点前打卡5次）
func AchievementCheckEarlyBird(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现早上6点前打卡次数统计
	// 暂时使用打卡总次数作为判断条件
	if user.Daka >= 5 {
		err := repository.UpdateAchievementHadDone(userID, "早起鸟")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "早起鸟"})
			return
		}
	}
}

// 成就检测：夜猫子（晚上10点后打卡5次）
func AchievementCheckNightOwl(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现晚上10点后打卡次数统计
	// 暂时使用打卡总次数作为判断条件
	if user.Daka >= 5 {
		err := repository.UpdateAchievementHadDone(userID, "夜猫子")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "夜猫子"})
			return
		}
	}
}

// 成就检测：完美主义（连续10次满分完成flag）
func AchievementCheckPerfectStreak(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现连续满分完成flag次数统计
	// 暂时使用完成flag总数作为判断条件
	if user.FlagNumber >= 10 {
		err := repository.UpdateAchievementHadDone(userID, "完美主义")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "完美主义"})
			return
		}
	}
}

// 成就检测：全能选手（完成5种不同标签的flag）
func AchievementCheckAllRounder(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现不同标签flag统计
	// 暂时使用完成flag总数作为判断条件
	if user.FlagNumber >= 5 {
		err := repository.UpdateAchievementHadDone(userID, "全能选手")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "全能选手"})
			return
		}
	}
}

// 成就检测：学习狂人（累计学习时间超过5000分钟）
func AchievementCheckLearn5000Min(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	var totalLearnTime int
	for _, learnTime := range user.LearnTimes {
		totalLearnTime += learnTime.Duration
	}
	if totalLearnTime >= 5000 {
		err := repository.UpdateAchievementHadDone(userID, "学习狂人")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "学习狂人"})
			return
		}
	}
}

// 成就检测：社交达人（发布10条动态）
func AchievementCheckPost10Times(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现动态发布次数统计
	// 暂时使用flag总数作为判断条件
	if user.FlagNumber >= 10 {
		err := repository.UpdateAchievementHadDone(userID, "社交达人")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "社交达人"})
			return
		}
	}
}

// 成就检测：时间管理者（连续30天完成至少1个flag）
func AchievementCheckDailyFlag30Days(userID uint) {
	user, err := repository.GetUserByID(userID)
	if err != nil {
		utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID})
		return
	}
	// TODO: 实现连续30天完成flag的检测逻辑
	// 暂时使用打卡总次数作为判断条件
	if user.Daka >= 30 {
		err := repository.UpdateAchievementHadDone(userID, "时间管理者")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "时间管理者"})
			return
		}
	}
}

// 成就检测：成就收集者（解锁10个徽章）
func AchievementCheckUnlock10Badges(userID uint) {
	achievements, err := repository.GetAchievementsByUserID(userID)
	if err != nil {
		utils.LogError("获取用户成就失败", logrus.Fields{"user_id": userID})
		return
	}
	unlockedCount := 0
	for _, achievement := range achievements {
		if achievement.HadDone {
			unlockedCount++
		}
	}
	if unlockedCount >= 10 {
		err := repository.UpdateAchievementHadDone(userID, "成就收集者")
		if err != nil {
			utils.LogError("更新成就状态失败", logrus.Fields{"user_id": userID, "achievement": "成就收集者"})
			return
		}
	}
}
