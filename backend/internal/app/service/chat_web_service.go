package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID        uint            `json:"id"`
	Conn      *websocket.Conn `json:"-"`
	Send      chan []byte     `json:"-"`
	CreatedAt time.Time       `json:"created_at"`
	RoomID    string          `json:"room_id"`
}

type Message struct {
	FromID     uint      `json:"from"`
	ToID       uint      `json:"to"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	RoomID     string    `json:"room_id"`
	UserName   string    `json:"user_name"`
	UserAvatar string    `json:"user_avatar"`
}

type ChatRoom struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	CreatorID  uint             `json:"creator_id"`
	Clients    map[uint]*Client `json:"-"`
	CreatedAt  time.Time        `json:"created_at"`
	LastActive time.Time        `json:"last_active"`
	MaxUsers   int              `json:"max_users"`
}

type Manager struct {
	mu            sync.RWMutex
	Rooms         map[string]*ChatRoom
	GlobalClients map[uint]*Client // 全局客户端映射，用于私聊
	Register      chan *Client
	Unregister    chan *Client
	Broadcast     chan Message
}

var manager = NewManager()

func init() {
	// 创建默认的3个聊天室
	defaultRooms := []struct {
		id   string
		name string
	}{
		{"room-1", "学习交流室"},
		{"room-2", "休闲娱乐室"},
		{"room-3", "技术讨论室"},
	}

	for _, room := range defaultRooms {
		manager.Rooms[room.id] = &ChatRoom{
			ID:         room.id,
			Name:       room.name,
			CreatorID:  0, // 系统创建
			Clients:    make(map[uint]*Client),
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
			MaxUsers:   50,
		}
	}

	go manager.Start()
	go manager.CleanupEmptyRooms() // 定期清理空房间
}

// WebSocket处理函数
func WsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.LogInfo("WebSocket连接请求到达", nil)
		// 从 JWT 中间件获取用户 ID
		id, ok := getCurrentUserID(c)
		if !ok || id == 0 {
			utils.LogError("WebSocket用户ID验证失败", nil)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权或 token 无效"})
			return
		}

		// 获取房间ID参数（私聊时可以不提供）
		roomID := c.Query("room_id")

		utils.LogInfo("WebSocket用户ID验证成功", map[string]interface{}{"user_id": id, "room_id": roomID})

		// 如果提供了房间ID，检查房间是否存在和是否已满
		if roomID != "" {
			manager.mu.RLock()
			room, exists := manager.Rooms[roomID]
			manager.mu.RUnlock()

			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
				return
			}

			// 检查房间人数限制
			manager.mu.RLock()
			roomFull := len(room.Clients) >= room.MaxUsers
			manager.mu.RUnlock()

			if roomFull {
				c.JSON(http.StatusForbidden, gin.H{"error": "房间已满"})
				return
			}
		}

		// 升级为 WebSocket 连接
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			utils.LogError("WebSocket连接升级失败", map[string]interface{}{"error": err.Error()})
			return
		}

		client := &Client{
			ID:        id,
			Conn:      conn,
			Send:      make(chan []byte, 256),
			RoomID:    roomID,
			CreatedAt: time.Now(),
		}

		manager.Register <- client
		utils.LogInfo("✅ WebSocket连接成功", map[string]interface{}{"user_id": id, "room_id": roomID, "remote_addr": c.Request.RemoteAddr})

		//埋点
		repository.AddTrackPointToDB(id, "用户使用聊天功能")

		go ReadPump(client)
		go WritePump(client)
	}
}

// 创建新的管理器
func NewManager() *Manager {
	return &Manager{
		Rooms:         make(map[string]*ChatRoom),
		GlobalClients: make(map[uint]*Client),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Broadcast:     make(chan Message),
	}
}

// 启动管理器
func (manager *Manager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.mu.Lock()
			// 添加到全局客户端映射（用于私聊）
			manager.GlobalClients[client.ID] = client

			// 如果有房间ID，添加到房间
			if client.RoomID != "" {
				if room, ok := manager.Rooms[client.RoomID]; ok {
					room.Clients[client.ID] = client
					room.LastActive = time.Now()
					log.Printf("User %d connected to room %s (total: %d users)", client.ID, client.RoomID, len(room.Clients))
				}
			} else {
				log.Printf("User %d connected for private chat", client.ID)
			}
			manager.mu.Unlock()

		case client := <-manager.Unregister:
			manager.mu.Lock()
			// 从全局映射移除
			if _, exists := manager.GlobalClients[client.ID]; exists {
				delete(manager.GlobalClients, client.ID)
				close(client.Send)
			}

			// 如果在房间中，从房间移除
			if client.RoomID != "" {
				if room, ok := manager.Rooms[client.RoomID]; ok {
					if _, exists := room.Clients[client.ID]; exists {
						delete(room.Clients, client.ID)
						room.LastActive = time.Now()
						log.Printf("User %d disconnected from room %s (remaining: %d users)", client.ID, client.RoomID, len(room.Clients))
					}
				}
			}
			manager.mu.Unlock()

		case message := <-manager.Broadcast:
			manager.mu.RLock()
			data, _ := json.Marshal(message)

			// 私聊消息（to > 0）
			if message.ToID > 0 {
				// 发给目标用户
				if targetClient, ok := manager.GlobalClients[message.ToID]; ok {
					select {
					case targetClient.Send <- data:
						log.Printf("📨 Private message from %d to %d delivered", message.FromID, message.ToID)
					default:
						log.Printf("❌ Failed to send private message from %d to %d", message.FromID, message.ToID)
					}
				} else {
					log.Printf("⚠️ Target user %d not online for private message", message.ToID)
				}

				// 也发回给发送者（让发送者看到确认，并且刷新后能看到历史）
				if senderClient, ok := manager.GlobalClients[message.FromID]; ok {
					select {
					case senderClient.Send <- data:
						log.Printf("✅ Private message echoed back to sender %d", message.FromID)
					default:
						log.Printf("⚠️ Failed to echo private message to sender %d", message.FromID)
					}
				}
			} else if message.RoomID != "" {
				// 房间广播消息（发给所有人，包括发送者）
				if room, ok := manager.Rooms[message.RoomID]; ok {
					room.LastActive = time.Now()
					successCount := 0
					for clientID, client := range room.Clients {
						select {
						case client.Send <- data:
							successCount++
						default:
							log.Printf("⚠️ Failed to send to client %d in room %s", clientID, message.RoomID)
							close(client.Send)
							delete(room.Clients, client.ID)
						}
					}
					log.Printf("📢 Room message broadcast: from=%d room=%s sent_to=%d/%d users", message.FromID, message.RoomID, successCount, len(room.Clients))
				}
			}
			manager.mu.RUnlock()
		}
	}
}

// 清理空房间（10小时无人则删除，默认房间除外）
func (manager *Manager) CleanupEmptyRooms() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		manager.mu.Lock()
		now := time.Now()
		for roomID, room := range manager.Rooms {
			// 跳过默认房间
			if roomID == "room-1" || roomID == "room-2" || roomID == "room-3" {
				continue
			}
			// 如果房间为空且超过10小时无活动，删除房间
			if len(room.Clients) == 0 && now.Sub(room.LastActive) > 10*time.Hour {
				delete(manager.Rooms, roomID)
				utils.LogInfo("删除空闲聊天室", logrus.Fields{"room_id": roomID, "room_name": room.Name})
			}
		}
		manager.mu.Unlock()
	}
}

// 从前端读取信息
func ReadPump(client *Client) {
	defer func() {
		manager.Unregister <- client
		client.Conn.Close()
	}()
	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			utils.LogError("WebSocket读取消息失败", map[string]interface{}{"user_id": client.ID, "error": err.Error()})
			break
		}
		message := Message{}
		err = json.Unmarshal(data, &message)
		if err != nil {
			utils.LogError("WebSocket消息解析失败", map[string]interface{}{"user_id": client.ID, "error": err.Error()})
			continue
		}
		message.FromID = client.ID
		message.RoomID = client.RoomID
		message.CreatedAt = time.Now()

		// 获取发送者用户信息
		user, err := repository.GetUserByID(client.ID)
		if err == nil {
			message.UserName = user.Name
			if user.HeadShow > 0 && user.HeadShow <= 6 {
				avatarFiles := []string{"131601", "131629", "131937", "131951", "132014", "133459"}
				message.UserAvatar = "/src/assets/images/screenshot_20251114_" + avatarFiles[user.HeadShow-1] + ".png"
			}
		}

		// 保存消息到数据库
		chatMsg := model.ChatMessage{
			FromUserID: message.FromID,
			ToUserID:   message.ToID,
			RoomID:     message.RoomID,
			Content:    message.Content,
			CreatedAt:  message.CreatedAt,
		}
		err = repository.SaveChatMessage(&chatMsg)
		if err != nil {
			utils.LogError("保存聊天消息失败", map[string]interface{}{"error": err.Error(), "from": message.FromID, "to": message.ToID})
		} else {
			utils.LogInfo("💾 消息已保存", map[string]interface{}{"id": chatMsg.ID, "from": message.FromID, "to": message.ToID, "room": message.RoomID, "content": message.Content})
		}

		manager.Broadcast <- message
	}
}

// 向前端写信息
func WritePump(client *Client) {
	defer client.Conn.Close()
	for message := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

// 获取聊天室列表
func GetChatRooms() gin.HandlerFunc {
	return func(c *gin.Context) {
		manager.mu.RLock()
		defer manager.mu.RUnlock()

		type RoomInfo struct {
			ID        string    `json:"id"`
			Name      string    `json:"name"`
			UserCount int       `json:"user_count"`
			MaxUsers  int       `json:"max_users"`
			CreatedAt time.Time `json:"created_at"`
			CreatorID uint      `json:"creator_id"`
		}

		rooms := make([]RoomInfo, 0)
		for _, room := range manager.Rooms {
			rooms = append(rooms, RoomInfo{
				ID:        room.ID,
				Name:      room.Name,
				UserCount: len(room.Clients),
				MaxUsers:  room.MaxUsers,
				CreatedAt: room.CreatedAt,
				CreatorID: room.CreatorID,
			})
		}

		c.JSON(http.StatusOK, gin.H{"rooms": rooms})
	}
}

// 创建聊天室
func CreateChatRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "房间名称不能为空"})
			return
		}

		userID, _ := getCurrentUserID(c)

		manager.mu.Lock()
		defer manager.mu.Unlock()

		// 检查聊天室数量限制
		if len(manager.Rooms) >= 10 {
			c.JSON(http.StatusForbidden, gin.H{"error": "聊天室数量已达上限（最多10个）"})
			return
		}

		// 生成房间ID
		roomID := "room-" + time.Now().Format("20060102150405")

		room := &ChatRoom{
			ID:         roomID,
			Name:       req.Name,
			CreatorID:  userID,
			Clients:    make(map[uint]*Client),
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
			MaxUsers:   50,
		}

		manager.Rooms[roomID] = room
		utils.LogInfo("创建聊天室成功", logrus.Fields{"room_id": roomID, "name": req.Name, "creator_id": userID})

		c.JSON(http.StatusOK, gin.H{
			"id":         room.ID,
			"name":       room.Name,
			"creator_id": room.CreatorID,
			"created_at": room.CreatedAt,
			"max_users":  room.MaxUsers,
		})
	}
}

// 删除聊天室（仅创建者或系统可删除）
func DeleteChatRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("room_id")
		userID, _ := getCurrentUserID(c)

		// 不能删除默认房间
		if roomID == "room-1" || roomID == "room-2" || roomID == "room-3" {
			c.JSON(http.StatusForbidden, gin.H{"error": "不能删除默认聊天室"})
			return
		}

		manager.mu.Lock()
		defer manager.mu.Unlock()

		room, exists := manager.Rooms[roomID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "聊天室不存在"})
			return
		}

		// 检查权限
		if room.CreatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "只有创建者可以删除聊天室"})
			return
		}

		// 关闭所有连接
		for _, client := range room.Clients {
			close(client.Send)
			client.Conn.Close()
		}

		delete(manager.Rooms, roomID)
		utils.LogInfo("删除聊天室", logrus.Fields{"room_id": roomID, "name": room.Name})

		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}

// 获取聊天室历史消息
func GetChatHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("room_id")
		limit := 30
		if limitParam := c.Query("limit"); limitParam != "" {
			var l int
			if _, err := fmt.Sscanf(limitParam, "%d", &l); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		messages, err := repository.GetChatHistory(roomID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史消息失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}

// 获取私聊历史消息
func GetPrivateChatHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}

		targetUserID := c.Query("target_user_id")
		if targetUserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少目标用户ID"})
			return
		}

		var targetID uint
		if _, err := fmt.Sscanf(targetUserID, "%d", &targetID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}

		limit := 30
		if limitParam := c.Query("limit"); limitParam != "" {
			var l int
			if _, err := fmt.Sscanf(limitParam, "%d", &l); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		messages, err := repository.GetPrivateChatHistory(userID, targetID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史消息失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}

// 获取私聊会话列表（按对方用户分组，显示最新消息）
func GetPrivateConversations() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}

		utils.LogInfo("获取私聊会话列表", logrus.Fields{"user_id": userID})

		conversations, err := repository.GetPrivateConversations(userID)
		if err != nil {
			utils.LogError("获取私聊会话列表失败", logrus.Fields{"user_id": userID, "error": err.Error()})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会话列表失败"})
			return
		}

		// 确保返回空数组而不是null
		if conversations == nil {
			conversations = []repository.Conversation{}
		}

		utils.LogInfo("会话列表获取成功", logrus.Fields{"user_id": userID, "count": len(conversations)})
		c.JSON(http.StatusOK, gin.H{"conversations": conversations})
	}
}
