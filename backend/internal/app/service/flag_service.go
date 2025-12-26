package service

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 获取用户flag
func GetUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		flags, err := repository.GetFlagsByUserID(id)
		if err != nil {
			c.JSON(401, gin.H{"error": "获取flag失败,请重新再试..."})
			log.Print("Get flags error")
			return
		}
		utils.LogInfo("获取用户flag成功", logrus.Fields{"user_id": id, "flag_count": len(flags)})
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}

// 添加用户flag
func PostUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var flag struct {
			Title  string `json:"title"`
			Detail string `json:"detail"`

			Label              int    `json:"label"`    // 前端发送数字1-5
			Priority           int    `json:"priority"` // 前端发送数字1-4
			Total              int    `json:"total"`
			Points             int    `json:"points"`
			DailyLimit         int    `json:"dailyLimit"`         // 每日完成次数限制
			IsRecurring        bool   `json:"isRecurring"`        // 是否循环任务
			EndTime            string `json:"endTime"`            // 改为string，手动解析
			StartTime          string `json:"startTime"`          // 改为string，手动解析
			ReminderTime       string `json:"reminderTime"`       // 提醒时间 (HH:MM 格式)
			EnableNotification bool   `json:"enableNotification"` // 是否启用提醒
		}
		if err := c.ShouldBindJSON(&flag); err != nil {
			c.JSON(500, gin.H{"err": "添加flag失败,请重新再试..."})
			log.Printf("Binding error: %v", err)
			return
		}

		// 验证label范围(1-5)，设置默认值
		if flag.Label < 1 || flag.Label > 5 {
			log.Printf("⚠️ Invalid label: %d, defaulting to 1", flag.Label)
			flag.Label = 1 // 默认学习类
		}

		// 验证priority范围(1-4)，设置默认值
		if flag.Priority < 1 || flag.Priority > 4 {
			log.Printf("⚠️ Invalid priority: %d, defaulting to 3", flag.Priority)
			flag.Priority = 3 // 默认一般
		}

		// 验证daily_limit，设置默认值
		if flag.DailyLimit < 1 {
			flag.DailyLimit = 1 // 默认每天至少1次
		}

		// 验证total，设置默认值
		if flag.Total < 1 {
			flag.Total = 1
		}

		// 解析时间字符串，只保留年月日，时分秒设为00:00:00
		// 如果前端不传日期（空字符串），则使用 nil（数据库存为 NULL，表示每天）
		var startTime *time.Time
		if flag.StartTime != "" {
			parsedStart, parseErr := time.Parse(time.RFC3339, flag.StartTime)
			if parseErr != nil {
				log.Printf("⚠️ 解析起始日期失败: %v, 使用 NULL（每天）", parseErr)
				startTime = nil // NULL，表示每天
			} else {
				t := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, parsedStart.Location())
				startTime = &t
			}
		}

		var endTime *time.Time
		if flag.EndTime != "" {
			parsedEnd, parseErr := time.Parse(time.RFC3339, flag.EndTime)
			if parseErr != nil {
				log.Printf("⚠️ 解析结束日期失败: %v, 使用 NULL（每天）", parseErr)
				endTime = nil // NULL，表示每天
			} else {
				t := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 23, 59, 59, 0, parsedEnd.Location())
				endTime = &t
			}
		}

		flag_model := model.Flag{
			Title:              flag.Title,
			Detail:             flag.Detail,
			Label:              flag.Label,
			Priority:           flag.Priority,
			DailyTotal:         flag.Total, // 每日所需完成次数
			Points:             flag.Points,
			CreatedAt:          time.Now(),
			StartTime:          startTime,
			EndTime:            endTime,
			ReminderTime:       flag.ReminderTime,       // 提醒时间
			EnableNotification: flag.EnableNotification, // 是否启用提醒
		}
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(402, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		err := repository.AddFlagToDB(id, flag_model)
		if err != nil {
			c.JSON(400, gin.H{"error": "添加flag失败,请重新再试..."})
			utils.LogError("数据库添加flag失败", logrus.Fields{})
			return
		}
		utils.LogInfo("添加用户flag成功", logrus.Fields{"user_id": id, "flag": flag.Title})
		// 重新查询以获取自动生成的ID
		flags, _ := repository.GetFlagsByUserID(id)
		var createdFlag model.Flag
		if len(flags) > 0 {
			createdFlag = flags[len(flags)-1] // 最后一个是刚创建的
		} else {
			createdFlag = flag_model
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Flag创建成功",
			"flag":    createdFlag,
		})
	}
}

// 打卡用户flag
func DoneUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先读取原始body用于调试
		bodyBytes, _ := c.GetRawData()
		log.Printf("DoneUserFlags received body: %s", string(bodyBytes))
		// 重新设置body供后续读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req struct {
			ID uint `json:"id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误,请重新再试..."})
			log.Printf("DoneUserFlags Binding error: %v", err)
			return
		}

		log.Printf("DoneUserFlags parsed ID: %d", req.ID)

		if req.ID == 0 {
			c.JSON(400, gin.H{"error": "无效的flag ID"})
			log.Printf("DoneUserFlags: Invalid ID (0)")
			return
		}

		durtion := time.Now()
		id, _ := utils.GetCurrentUserID(c)
		if err := repository.UpdateUserDoFlag(id, durtion); err != nil {
			c.JSON(400, gin.H{"error": "打卡失败,请重新再试..."})
			return
		}
		flag, err := repository.GetFlagByID(req.ID)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			return
		}

		// 校验flag是否在有效日期范围内
		today := time.Now()
		if flag.StartTime != nil && today.Before(*flag.StartTime) {
			c.JSON(400, gin.H{"error": "该flag未到起始日期，无法打卡"})
			utils.LogInfo("打卡失败：未到起始日期", logrus.Fields{"flag_id": req.ID, "start_time": flag.StartTime})
			return
		}
		if flag.EndTime != nil && today.After(*flag.EndTime) {
			c.JSON(400, gin.H{"error": "该flag已过结束日期，无法打卡"})
			utils.LogInfo("打卡失败：已过结束日期", logrus.Fields{"flag_id": req.ID, "end_time": flag.EndTime})
			return
		}

		flag.Count += 1
		err = repository.UpdateFlagDoneNumber(req.ID, flag.Count)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag失败", logrus.Fields{})
			return
		}

		// 检查Flag是否完成
		if flag.Count >= flag.DailyTotal && !flag.Completed {
			// 标记Flag为已完成
			err = repository.UpdateFlagHadDone(req.ID, true)
			if err != nil {
				utils.LogError("更新Flag完成状态失败", logrus.Fields{"flag_id": req.ID, "error": err.Error()})
			}

			// 🔧 新增：Flag完成时自动禁用提醒（避免已完成flag占用提醒上限）
			if flag.EnableNotification {
				err = repository.UpdateFlagNotification(req.ID, false)
				if err != nil {
					utils.LogError("自动禁用已完成Flag的提醒失败", logrus.Fields{"flag_id": req.ID, "error": err.Error()})
				} else {
					utils.LogInfo("已完成Flag的提醒已自动禁用", logrus.Fields{"flag_id": req.ID})
				}
			}

			// 更新用户的完成Flag计数
			user, err := repository.GetUserByID(id)
			if err == nil {
				newFlagNumber := user.FlagNumber + 1
				err = repository.FlagNumberAddDB(id, newFlagNumber)
				if err != nil {
					utils.LogError("更新用户Flag计数失败", logrus.Fields{"user_id": id, "error": err.Error()})
				} else {
					utils.LogInfo("用户完成Flag，计数已更新", logrus.Fields{"user_id": id, "flag_id": req.ID, "new_count": newFlagNumber})
				}

				// 🔧 新增：自动增加积分（根据Flag积分字段）
				if flag.Points > 0 {
					err = repository.CountAddDB(id, flag.Points)
					if err != nil {
						utils.LogError("更新用户积分失败", logrus.Fields{"user_id": id, "error": err.Error()})
					} else {
						utils.LogInfo("用户完成Flag，积分已增加", logrus.Fields{"user_id": id, "flag_id": req.ID, "points": flag.Points})
					}
				}
			}
		}

		utils.LogInfo("用户打卡成功", logrus.Fields{"user_id": id, "flag_id": req.ID, "count": flag.Count, "total": flag.DailyTotal})
		c.JSON(200, gin.H{"success": true,
			"count": flag.Count, "completed": flag.Count >= flag.DailyTotal})
	}
}

// 删除flag
func DeleteUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "删除flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		err := repository.DeleteFlagFromDB(req.ID)
		if err != nil {
			c.JSON(400, gin.H{"error": "删除flag失败,请重新再试..."})
			utils.LogError("数据库删除flag失败", logrus.Fields{})
			return
		}
		utils.LogInfo("删除用户flag成功", logrus.Fields{"flag_id": req.ID})
		c.JSON(200, gin.H{"success": true})
	}
}

// 完成flag（旧接口，已废弃 - 建议使用doneFlag）
// ⚠️ 注意：此函数有bug，level参数用途不明，且积分计算错误
func FinshDoneFlag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id"`
		}
		level := c.Query("level")
		id, _ := utils.GetCurrentUserID(c)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		user, _ := repository.GetUserByID(id)
		flag, _ := repository.GetFlagByID(req.ID)
		// 将数字label转换为字符串保存
		labelMap := map[int]string{
			1: "生活",
			2: "学习",
			3: "工作",
			4: "兴趣",
			5: "运动",
		}
		labelStr := labelMap[flag.Label]
		if labelStr == "" {
			labelStr = "学习"
		}
		repository.SaveLabelToDB(id, labelStr)
		user.FlagNumber++
		repository.SaveUserToDB(user)

		// 🔧 修复：应该使用Flag的Points字段而不是level参数
		// level参数用途不明，改用flag.Points
		pointsToAdd := flag.Points
		if pointsToAdd > 0 {
			err := repository.CountAddDB(id, pointsToAdd)
			if err != nil {
				log.Printf("[error] 积分更新失败: %v", err)
			} else {
				log.Printf("[info] 用户完成Flag，积分已增加 - 用户ID: %d, Flag ID: %d, 积分: %d", id, req.ID, pointsToAdd)
			}
		}

		// 更新Flag完成数
		repository.FlagNumberAddDB(id, user.FlagNumber+1)

		err := repository.UpdateFlagHadDone(req.ID, true)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag完成状态失败", logrus.Fields{})
			return
		}
		utils.LogInfo("flag完成状态更新成功（旧接口）", logrus.Fields{"user_id": id, "flag_id": req.ID, "level": level})
		c.JSON(200, gin.H{"success": true})
	}
}

// 获取最新打卡的十个人
func GetRecentDoFlagUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetRecentDoneFlags()
		if err != nil {
			c.JSON(400, gin.H{"error": "获取最近打卡用户失败,请重新再试..."})
			utils.LogError("数据库获取最近打卡用户失败", logrus.Fields{})
			return
		}
		utils.LogInfo("获取最近打卡用户成功", logrus.Fields{"user_count": len(users)})
		c.JSON(200, gin.H{"users": users})
	}
}

// 获取已完成flag
func GetDoneFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		flags, err := repository.GetDoneFlagsByUserID(id)
		if err != nil {
			c.JSON(401, gin.H{"error": "获取已完成flag失败,请重新再试..."})
			utils.LogError("获取已完成flag失败", logrus.Fields{})
			return
		}
		utils.LogInfo("获取已完成flag成功", logrus.Fields{"user_id": id, "flag_count": len(flags)})
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}

// 获取未完成的完成flag
func GetNotDoneFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		flags, err := repository.GetUndoneFlagsByUserID(id)
		if err != nil {
			c.JSON(401, gin.H{"error": "获取未完成flag失败,请重新再试..."})
			utils.LogError("获取未完成flag失败", logrus.Fields{})
			return
		}
		utils.LogInfo("获取未完成flag成功", logrus.Fields{"user_id": id, "flag_count": len(flags)})
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}

// 切换flag公开状态
func UpdateFlagHide() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		// 该接口已废弃：分享状态由post_id控制，不再需要单独的可见性字段
		// 前端应通过创建/删除Post来控制分享状态
		utils.LogInfo("该接口已废弃，请使用Post相关接口", logrus.Fields{"flag_id": req.ID})
		c.JSON(200, gin.H{"success": true, "message": "该接口已废弃"})
	}
}

// 更新flag完整信息
func UpdateFlagInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID                 uint   `json:"id"`
			Title              string `json:"title"`
			Detail             string `json:"detail"`
			Label              int    `json:"label"`
			Priority           int    `json:"priority"`
			Total              int    `json:"total"`
			StartDate          string `json:"startDate"`
			EndDate            string `json:"endDate"`
			ReminderTime       string `json:"reminderTime"`
			EnableNotification bool   `json:"enableNotification"`
			PostID             *uint  `json:"postId"` // 添加PostID字段，保留分享状态
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数解析失败"})
			utils.LogError("更新Flag参数绑定失败", logrus.Fields{"error": err.Error()})
			return
		}

		// 验证flag是否存在
		_, err := repository.GetFlagByID(req.ID)
		if err != nil {
			c.JSON(400, gin.H{"error": "Flag不存在"})
			utils.LogError("Flag不存在", logrus.Fields{"flag_id": req.ID})
			return
		}

		// 构建更新数据
		updates := map[string]interface{}{
			"flag":                req.Title,
			"plan_content":        req.Detail,
			"label":               req.Label,
			"priority":            req.Priority,
			"daily_total":         req.Total,
			"reminder_time":       req.ReminderTime,
			"enable_notification": req.EnableNotification,
		}

		// 添加post_id字段（如果前端传递了）
		if req.PostID != nil {
			updates["post_id"] = req.PostID
		}

		// 添加可选的起始/结束时间
		if req.StartDate != "" {
			if startTime, err := time.Parse("2006-01-02", req.StartDate); err == nil {
				updates["start_time"] = startTime
			}
		}
		if req.EndDate != "" {
			if endTime, err := time.Parse("2006-01-02", req.EndDate); err == nil {
				updates["end_time"] = endTime
			}
		}

		utils.LogInfo("准备更新Flag", logrus.Fields{
			"flag_id": req.ID,
			"updates": updates,
		})

		err = repository.UpdateFlag(req.ID, updates)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag失败", logrus.Fields{"flag_id": req.ID, "error": err.Error()})
			return
		}

		utils.LogInfo("flag更新成功", logrus.Fields{"flag_id": req.ID})
		c.JSON(200, gin.H{"success": true, "message": "Flag更新成功"})
	}
}

// 获取有起始日期的flag（用于日历高亮）
func GetFlagsWithDates() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败"})
			return
		}
		today := time.Now()
		flags, err := repository.GetFlagsWithDatesByUserID(id, today)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取flag失败"})
			utils.LogError("获取有日期的flag失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}
		utils.LogInfo("获取有日期的flag成功", logrus.Fields{"user_id": id, "count": len(flags)})
		c.JSON(200, gin.H{"flags": flags})
	}
}

// 获取预设flag（未到起始日期的flag）
func GetPresetFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败"})
			return
		}
		today := time.Now()
		flags, err := repository.GetPresetFlagsByUserID(id, today)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取预设flag失败"})
			utils.LogError("获取预设flag失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}
		utils.LogInfo("获取预设flag成功", logrus.Fields{"user_id": id, "count": len(flags)})
		c.JSON(200, gin.H{"flags": flags})
	}
}

// 获取过期flag
func GetExpiredFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败"})
			return
		}
		today := time.Now()
		flags, err := repository.GetExpiredFlagsByUserID(id, today)
		if err != nil {
			c.JSON(500, gin.H{"error": "获取过期flag失败"})
			utils.LogError("获取过期flag失败", logrus.Fields{"user_id": id, "error": err.Error()})
			return
		}
		utils.LogInfo("获取过期flag成功", logrus.Fields{"user_id": id, "count": len(flags)})
		c.JSON(200, gin.H{"flags": flags})
	}
}

// 切换flag提醒状态（最多3个flag可以提醒）
func ToggleFlagNotification() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FlagID             uint `json:"flagId"`
			EnableNotification bool `json:"enableNotification"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "请求参数错误"})
			utils.LogError("切换flag提醒参数绑定错误", logrus.Fields{"error": err.Error()})
			return
		}

		userID, ok := utils.GetCurrentUserID(c)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败"})
			return
		}

		// 如果要启用提醒，检查当前已启用提醒的flag数量
		if req.EnableNotification {
			count, err := repository.CountEnabledNotificationFlags(userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "检查提醒数量失败"})
				utils.LogError("检查提醒数量失败", logrus.Fields{"user_id": userID, "error": err.Error()})
				return
			}
			if count >= 5 {
				c.JSON(400, gin.H{"error": "最多只能同时为5个flag启用提醒"})
				utils.LogWarn("flag提醒数量已达上限", logrus.Fields{"user_id": userID, "count": count})
				return
			}

			// 自动开启用户级Flag提醒总开关（如果还未开启）
			user, err := repository.GetUserByID(userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "获取用户信息失败"})
				utils.LogError("获取用户信息失败", logrus.Fields{"user_id": userID, "error": err.Error()})
				return
			}
			if !user.IsFlagRemind {
				err = repository.UpdateUserFlagRemindStatus(userID, true)
				if err != nil {
					c.JSON(500, gin.H{"error": "自动开启用户Flag提醒总开关失败"})
					utils.LogError("自动开启用户Flag提醒总开关失败", logrus.Fields{"user_id": userID, "error": err.Error()})
					return
				}
				utils.LogInfo("自动开启用户Flag提醒总开关", logrus.Fields{"user_id": userID})
			}
		}

		// 更新flag的提醒状态
		err := repository.UpdateFlagNotification(req.FlagID, req.EnableNotification)
		if err != nil {
			c.JSON(500, gin.H{"error": "更新flag提醒状态失败"})
			utils.LogError("更新flag提醒状态失败", logrus.Fields{"flag_id": req.FlagID, "error": err.Error()})
			return
		}

		utils.LogInfo("flag提醒状态更新成功", logrus.Fields{"user_id": userID, "flag_id": req.FlagID, "enabled": req.EnableNotification})
		c.JSON(200, gin.H{"success": true, "enableNotification": req.EnableNotification})
	}
}

// RecalcFlagNumbers 重新计算所有用户的flag_number
func RecalcFlagNumbers() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.LogInfo("开始重新计算所有用户的flag_number", nil)

		c.JSON(200, gin.H{
			"success": true,
			"message": "此功能当前不可用，请联系管理员",
		})
	}
}
