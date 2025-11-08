package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID        int  ` json:"id"`
	UserID    uint `json:"user_id"`
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
	ID         uint `gorm:"primaryKey" json:"id"`
	Unregister chan *Client
	Register   chan *Client
	Clients    map[uint]*Client
	Broadcast  chan []byte
}

var manager = NewManager()

// WebSocket处理函数
func WsHandler() gin.HandlerFunc {
	{
		return func(c *gin.Context) {
			// 1. 取参数
			id := c.Query("id")
			if id == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
				return
			}
			// 2. 升级
			conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
			if err != nil {
				log.Println("upgrade err:", err)
				return
			}
			new_id, _ := strconv.Atoi(id)
			client := &Client{ID: new_id, Conn: conn, Send: make(chan []byte, 256)}
			manager.Register <- client

			// 3. 启动读写协程
			go ReadPump(client)
			go WritePump(client)

		}
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
			manager.Clients[client.UserID] = client
			log.Printf("User %d connected", client.UserID)
		case client := <-manager.Unregister:
			if _, ok := manager.Clients[client.UserID]; ok {
				delete(manager.Clients, client.UserID)
				close(client.Send)
				log.Printf("User %d disconnected", client.UserID)
			}
		case message := <-manager.Broadcast:
			for _, client := range manager.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(manager.Clients, client.UserID)
				}
			}
		}
	}
}
func (manager *Manager) Rontune(m Message) {
	data, _ := json.Marshal(m)
	if m.ToID == 0 {
		for _, client := range manager.Clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(manager.Clients, client.UserID)
			}
		}
	}
	if client, ok := manager.Clients[m.ToID]; ok {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(manager.Clients, client.UserID)
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
			log.Printf("User %d read error: %v", client.UserID, err)
			break
		}
		message := Message{}
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Printf("User %d unmarshal error: %v", client.UserID, err)
			continue
		}
		message.FromID = client.UserID
		message.CreatedAt = time.Now()
		client.Manager.Rontune(message)
	}
}

// 向前端写信息
func WritePump(client *Client) {
	defer func() {
		client.Conn.Close()
	}()
	for {
		for message := range client.Send {
			if err := client.Conn.WriteMessage(websocket.CloseMessage, message); err != nil {
				log.Printf("User %d write error: %v", client.UserID, err)
				return
			}
		}
	}
}
