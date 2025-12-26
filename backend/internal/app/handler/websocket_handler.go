package handler

import (
	"net/http"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/service"
	"github.com/gin-gonic/gin"
)

// 获取所有聊天室信息
func GetRoomsInfo(c *gin.Context) {
	rooms := service.GetRoomManager().GetAllRooms()
	c.JSON(http.StatusOK, gin.H{
		"message": "获取聊天室列表成功",
		"data":    rooms,
	})
}

// 获取房间内的用户列表
func GetRoomUsers(c *gin.Context) {
	roomID := c.Query("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少 roomId 参数",
		})
		return
	}

	users := service.GetRoomManager().GetRoomUsers(roomID)
	c.JSON(http.StatusOK, gin.H{
		"message": "获取房间用户列表成功",
		"roomId":  roomID,
		"users":   users,
	})
}
