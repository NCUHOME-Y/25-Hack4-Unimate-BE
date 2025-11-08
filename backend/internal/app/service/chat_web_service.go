package service

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID        uint ` json:"id"`
	Conn      *websocket.Conn
	Manager   *Manager    `json:"-"`
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
		log.Printf("[WebSocket] 收到连接请求 - RemoteAddr: %s", c.Request.RemoteAddr)

		// 从 JWT 中间件获取用户 ID
		id, ok := getCurrentUserID(c)
		if !ok || id == 0 {
			log.Printf("[WebSocket] 获取用户ID失败 - ok: %v, id: %d", ok, id)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权或 token 无效"})
			return
		}

		log.Printf("[WebSocket] 用户ID验证成功: %d, 准备升级连接", id)

		// 升级为 WebSocket 连接
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WebSocket] 升级失败 user=%d remote=%s err=%v", id, c.Request.RemoteAddr, err)
			return
		}

		client := &Client{ID: id, Conn: conn, Send: make(chan []byte, 256), Manager: manager}
		manager.Register <- client
		log.Printf("[WebSocket] ✅ 连接成功 user=%d remote=%s", id, c.Request.RemoteAddr)

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
		client.Manager.Unregister <- client
		client.Conn.Close()
	}()
	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("User %d read error: %v", client.ID, err)
			break
		}
		message := Message{}
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Printf("User %d unmarshal error: %v", client.ID, err)
			continue
		}
		message.FromID = client.ID
		message.CreatedAt = time.Now()
		client.Manager.Route(message)
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
