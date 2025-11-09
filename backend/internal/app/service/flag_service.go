package service

import (
	"Heckweek/internal/app/model"
	"Heckweek/internal/app/repository"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
			log.Print("Add flag to DB error")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "添加flag成功!"})
	}
}

// 打卡用户flag
func DoneUserFlags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID         uint `json:"id"`
			DoneNumber int  `json:"done_number"`
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
		err := repository.UpdateFlagDoneNumber(req.ID, req.DoneNumber)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			return
		}
		c.JSON(200, gin.H{"message": "打卡成功"})
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
			return
		}
		c.JSON(200, gin.H{"message": "删除flag成功"})
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
		count, _ := strconv.Atoi(level)
		newcount := user.Count + count
		err := repository.CountAddDB(id, newcount)
		if err != nil {
			log.Printf("[error] 积分更新失败: %v", err)
		}
		err = repository.UpdateFlagHadDone(req.ID)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			return
		}
		c.JSON(200, gin.H{"message": "flag完成状态更新成功"})
	}
}

// 获取最新打卡的十个人
func GetRecentDoFlagUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetRecentDoneFlags()
		if err != nil {
			c.JSON(400, gin.H{"error": "获取最近打卡用户失败,请重新再试..."})
			log.Print("Get recent do flag users error")
			return
		}
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
			log.Print("Get finished flags error")
			return
		}
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
			log.Print("Get not finished flags error")
			return
		}
		c.JSON(http.StatusOK, gin.H{"flags": flags})
	}
}
