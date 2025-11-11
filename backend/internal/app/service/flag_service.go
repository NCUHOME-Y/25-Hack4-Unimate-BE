package service

import (
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
			Flag           string    `json:"flag"`
			PlanContent    string    `json:"plan_content"`
			IsHiden        bool      `json:"is_hiden"`
			Label          string    `json:"label"`
			Priority       int       `json:"priority"`
			PlanDoneNumber int       `json:"plan_done_number"`
			DeadTime       time.Time `json:"deadtime"`
			StartTime      time.Time `json:"starttime"`
		}
		if err := c.ShouldBindJSON(&flag); err != nil {
			c.JSON(500, gin.H{"err": "添加flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		flag_model := model.Flag{
			Flag:           flag.Flag,
			PlanContent:    flag.PlanContent,
			IsHiden:        flag.IsHiden,
			Label:          flag.Label,
			Priority:       flag.Priority,
			PlanDoneNumber: flag.PlanDoneNumber,
			CreatedAt:      time.Now(),
			StartTime:      flag.StartTime,
			DeadTime:       flag.DeadTime,
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
		utils.LogInfo("添加用户flag成功", logrus.Fields{"user_id": id, "flag": flag.Flag})
		c.JSON(http.StatusOK, gin.H{"success": true,
			"flag": flag_model})
	}
}

// 打卡用户flag
func DoneUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"flag_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新flag失败,请重新再试..."})
			log.Print("Binding error")
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
		flag.DoneNumber += 1
		err = repository.UpdateFlagDoneNumber(req.ID, flag.DoneNumber)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag失败", logrus.Fields{})
			return
		}
		utils.LogInfo("用户打卡成功", logrus.Fields{"user_id": id, "flag_id": req.ID})
		c.JSON(200, gin.H{"success": true,
			"count": flag.DoneNumber})
	}
}

// 删除flag
func DeleteUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"flag_id"`
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
			ID uint `json:"flag_id"`
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
			ID uint `json:"flag_id"`
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
		flag.IsHiden = !flag.IsHiden
		err = repository.UpdateFlagVisibility(req.ID, flag.IsHiden)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			utils.LogError("数据库更新flag公开状态失败", logrus.Fields{})
			return
		}
		utils.LogInfo("flag公开状态更新成功", logrus.Fields{"flag_id": req.ID})
		c.JSON(200, gin.H{"success": true})
	}
}
