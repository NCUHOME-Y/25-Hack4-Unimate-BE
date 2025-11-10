package service

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	utils "github.com/NCUHOME-Y/25-Hack4-Unimate-BE/util"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID        uint ` json:"id"`
	Conn      *websocket.Conn
	Send      chan []byte `json:"-"`
	CreatedAt time.Time   `json:"created_at"`
}

type Message struct {
	FromID    uint      `json:"from"`
	ToID      uint      `json:"to"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Manager struct {
	mu         sync.Mutex
	ID         uint `gorm:"primaryKey" json:"id"`
	Unregister chan *Client
	Register   chan *Client
	Clients    map[uint]*Client
	Broadcast  chan []byte
}

var manager = NewManager()

func init() {
	go manager.Start()
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

		utils.LogInfo("WebSocket用户ID验证成功", map[string]interface{}{"user_id": id})

		// 升级为 WebSocket 连接
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			utils.LogError("WebSocket连接升级失败", map[string]interface{}{"error": err.Error()})
			return
		}

		client := &Client{ID: id, Conn: conn, Send: make(chan []byte, 256)}
		manager.Register <- client
		utils.LogInfo("✅ WebSocket连接成功", map[string]interface{}{"user_id": id, "remote_addr": c.Request.RemoteAddr})

		go ReadPump(client)
		go WritePump(client)
	}
}

// 创建新的管理器
func NewManager() *Manager {
	return &Manager{
		Unregister: make(chan *Client),
		Register:   make(chan *Client),
		Clients:    make(map[uint]*Client),
		Broadcast:  make(chan []byte),
	}
}

// 启动管理器
func (manager *Manager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.mu.Lock()
			manager.Clients[client.ID] = client
			manager.mu.Unlock()
			log.Printf("User %d connected", client.ID)
		case client := <-manager.Unregister:
			manager.mu.Lock()
			if _, ok := manager.Clients[client.ID]; ok {
				delete(manager.Clients, client.ID)
				close(client.Send)
				log.Printf("User %d disconnected", client.ID)
			}
			manager.mu.Unlock()
		case message := <-manager.Broadcast:
			manager.mu.Lock()
			for _, client := range manager.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(manager.Clients, client.ID)
				}
			}
			manager.mu.Unlock()
		}
	}
}
func (manager *Manager) Route(m Message) {
	data, _ := json.Marshal(m)
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if m.ToID == 0 { // 广播
		for _, client := range manager.Clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(manager.Clients, client.ID)
			}
		}
		return
	}
	if client, ok := manager.Clients[m.ToID]; ok {
		select {
		case client.Send <- data:

		default:
			close(client.Send)
			delete(manager.Clients, client.ID)
		}
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
		message.CreatedAt = time.Now()
		manager.Route(message)
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
