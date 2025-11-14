package service

import (
	"log"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"

	"strconv"

	"github.com/gin-gonic/gin"
)

// 积分函数
func AddUserCount(count string, id uint) {
	user, _ := repository.GetUserByID(id)
	var countInt, _ = strconv.Atoi(count)
	newcount := user.Count + countInt
	err := repository.CountAddDB(id, newcount)
	if err != nil {
		log.Printf("[error] 积分更新失败: %v", err)
		return
	}
}

// 积分排行榜
func GetUserCount() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetUserByCount()
		if err != nil {
			c.JSON(500, gin.H{"error": "获取排行榜失败,请重新再试..."})
			return
		}
		//埋点
		repository.AddTrackPointToDB(0, "查看积分排行榜")
		c.JSON(200, gin.H{"message": "获取排行榜成功", "data": users})
	}
}

// 月学习时间排行榜
func GetUserMonthLearnTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetUserByMonthLearnTime()
		if err != nil {
			c.JSON(500, gin.H{"error": "获取排行榜失败,请重新再试..."})
			return
		}
		repository.AddTrackPointToDB(0, "查看月学习时间排行榜")
		c.JSON(200, gin.H{"message": "获取排行榜成功", "data": users})
	}
}

// 总打卡数排行榜
func GetUserTotalDaka() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetUserByDaka()
		if err != nil {
			c.JSON(500, gin.H{"error": "获取排行榜失败,请重新再试..."})
		}
		repository.AddTrackPointToDB(0, "查看总打卡数排行榜")
		c.JSON(200, gin.H{"message": "获取排行榜成功", "data": users})
	}
}

// 按flag数量排序
func GetUserByFlagNumber() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repository.GetUserByFlagNumber()
		if err != nil {
			c.JSON(500, gin.H{"error": "获取排行榜失败,请重新再试..."})
		}
		repository.AddTrackPointToDB(0, "查看flag数量排行榜")
		c.JSON(200, gin.H{"message": "获取排行榜成功", "data": users})
	}
}
