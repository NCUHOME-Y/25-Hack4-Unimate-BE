package service

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
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
		id, ok := getCurrentUserID(c)
		log.Printf("[debug] user_id = %d", id)
		if !ok {
			c.JSON(400, gin.H{"error": "获取用户信息失败,请重新再试..."})
			return
		}
		flags, err := repository.GetFlagsByUserID(id)
		log.Printf("[debug] sql err=%v  len=%d", err, len(flags))
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
			Title       string `json:"title"`
			Detail      string `json:"detail"`
			IsPublic    bool   `json:"is_public"`
			Label       int    `json:"label"`    // 前端发送数字1-5
			Priority    int    `json:"priority"` // 前端发送数字1-4
			Total       int    `json:"total"`
			Points      int    `json:"points"`
			DailyLimit  int    `json:"daily_limit"`  // 每日完成次数限制
			IsRecurring bool   `json:"is_recurring"` // 是否循环任务
			EndTime     string `json:"end_time"`     // 改为string，手动解析
			StartTime   string `json:"start_time"`   // 改为string，手动解析
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
		// 如果前端不传日期（空字符串），则使用零值（表示无限期）
		var startTime time.Time
		if flag.StartTime != "" {
			parsedStart, parseErr := time.Parse(time.RFC3339, flag.StartTime)
			if parseErr != nil {
				log.Printf("⚠️ 解析起始日期失败: %v, 使用零值（无限期）", parseErr)
				startTime = time.Time{} // 零值，表示无限期
			} else {
				startTime = time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, parsedStart.Location())
			}
		}

		var endTime time.Time
		if flag.EndTime != "" {
			parsedEnd, parseErr := time.Parse(time.RFC3339, flag.EndTime)
			if parseErr != nil {
				log.Printf("⚠️ 解析结束日期失败: %v, 使用零值（无限期）", parseErr)
				endTime = time.Time{} // 零值，表示无限期
			} else {
				endTime = time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 23, 59, 59, 0, parsedEnd.Location())
			}
		}

		flag_model := model.Flag{
			Title:     flag.Title,
			Detail:    flag.Detail,
			IsPublic:  flag.IsPublic,
			Label:     flag.Label,
			Priority:  flag.Priority,
			Total:     flag.Total,
			Points:    flag.Points, // 添加积分字段
			CreatedAt: time.Now(),
			StartTime: startTime,
			EndTime:   endTime,
		}
		id, ok := getCurrentUserID(c)
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
		id, _ := getCurrentUserID(c)
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
		if !flag.StartTime.IsZero() && today.Before(flag.StartTime) {
			c.JSON(400, gin.H{"error": "该flag未到起始日期，无法打卡"})
			utils.LogInfo("打卡失败：未到起始日期", logrus.Fields{"flag_id": req.ID, "start_time": flag.StartTime})
			return
		}
		if !flag.EndTime.IsZero() && today.After(flag.EndTime) {
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
		if flag.Count >= flag.Total && !flag.Completed {
			// 标记Flag为已完成
			err = repository.UpdateFlagHadDone(req.ID, true)
			if err != nil {
				utils.LogError("更新Flag完成状态失败", logrus.Fields{"flag_id": req.ID, "error": err.Error()})
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

		utils.LogInfo("用户打卡成功", logrus.Fields{"user_id": id, "flag_id": req.ID, "count": flag.Count, "total": flag.Total})
		c.JSON(200, gin.H{"success": true,
			"count": flag.Count, "completed": flag.Count >= flag.Total})
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

// 完成flag
func FinshDoneFlag() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id"`
		}
		level := c.Query("level")
		id, _ := getCurrentUserID(c)
		log.Printf("[debug] user_id = %d", id)
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
		count, _ := strconv.Atoi(level)
		newcount := user.Count + count
		repository.FlagNumberAddDB(id, user.FlagNumber+1)
		err := repository.CountAddDB(id, newcount)
		if err != nil {
			log.Printf("[error] 积分更新失败: %v", err)
		}
		err = repository.UpdateFlagHadDone(req.ID, true)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag完成状态失败", logrus.Fields{})
			return
		}
		utils.LogInfo("flag完成状态更新成功", logrus.Fields{"user_id": id, "flag_id": req.ID})
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
		id, ok := getCurrentUserID(c)
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
		id, ok := getCurrentUserID(c)
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
		flag, err := repository.GetFlagByID(req.ID)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			return
		}
		flag.IsPublic = !flag.IsPublic
		err = repository.UpdateFlagVisibility(req.ID, flag.IsHidden)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag公开状态失败", logrus.Fields{})
			return
		}
		utils.LogInfo("flag公开状态更新成功", logrus.Fields{"flag_id": req.ID})
		c.JSON(200, gin.H{"success": true})
	}
}

// 更新flag完整信息
func UpdateFlagInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID       uint   `json:"id"`
			Title    string `json:"title"`
			Detail   string `json:"detail"`
			Label    int    `json:"label"`
			Priority int    `json:"priority"`
			Total    int    `json:"total"`
			IsPublic bool   `json:"is_public"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}

		// 验证flag是否存在
		_, err := repository.GetFlagByID(req.ID)
		if err != nil {
			c.JSON(400, gin.H{"error": "Flag不存在"})
			return
		}

		// 构建更新数据
		updates := map[string]interface{}{
			"title":     req.Title,
			"detail":    req.Detail,
			"label":     req.Label,
			"priority":  req.Priority,
			"total":     req.Total,
			"is_public": req.IsPublic,
		}

		err = repository.UpdateFlag(req.ID, updates)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag失败", logrus.Fields{"flag_id": req.ID})
			return
		}

		utils.LogInfo("flag更新成功", logrus.Fields{"flag_id": req.ID})
		c.JSON(200, gin.H{"success": true, "message": "Flag更新成功"})
	}
}

// 获取有起始日期的flag（用于日历高亮）
func GetFlagsWithDates() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := getCurrentUserID(c)
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
		id, ok := getCurrentUserID(c)
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
		id, ok := getCurrentUserID(c)
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
