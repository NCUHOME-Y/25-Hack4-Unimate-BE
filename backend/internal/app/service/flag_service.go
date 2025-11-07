package service

import (
	"Heckweek/internal/app/model"
	"Heckweek/internal/app/repository"
	"log"
	"net/http"
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
			Flag           string `json:"flag"`
			PlanContent    string `json:"plan_content"`
			IsHiden        bool   `json:"is_hiden"`
			PlanDoneNumber int    `json:"plan_done_number"`
			DeadTime       string `json:"deadtime"`
		}
		if err := c.ShouldBindJSON(&flag); err != nil {
			c.JSON(500, gin.H{"err": "添加flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		t, _ := time.Parse(flag.DeadTime, "2006-01-02 15:04:05")
		flag_model := model.Flag{
			Flag:           flag.Flag,
			PlanContent:    flag.PlanContent,
			IsHiden:        flag.IsHiden,
			PlanDoneNumber: flag.PlanDoneNumber,
			DeadTime:       t,
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

// 完成用户flag
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
			ID      uint `json:"id"`
			HadDone bool `json:"had_done"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(500, gin.H{"err": "更新flag失败,请重新再试..."})
			log.Print("Binding error")
			return
		}
		err := repository.UpdateFlagHadDone(req.ID, req.HadDone)
		if err != nil {
			c.JSON(400, gin.H{"error": "更新flag失败,请重新再试..."})
			return
		}
		c.JSON(200, gin.H{"message": "flag完成状态更新成功"})
	}
}
