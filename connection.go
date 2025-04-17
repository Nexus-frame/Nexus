package Nexus

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection 表示一个WebSocket连接
type Connection struct {
	ws         *websocket.Conn
	send       chan []byte
	engine     *Engine
	closed     bool
	closeMu    sync.Mutex
	closeChan  chan struct{}
	lastActive time.Time
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsHandler WebSocket处理函数类型
type WsHandler func(e *Engine, w http.ResponseWriter, r *http.Request)

// serveWs 处理WebSocket连接请求
func serveWs(e *Engine, w http.ResponseWriter, r *http.Request) {
	// 使用配置的参数更新upgrader
	upgrader.ReadBufferSize = e.config.WebSocketConfig.ReadBufferSize
	upgrader.WriteBufferSize = e.config.WebSocketConfig.WriteBufferSize
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return e.config.WebSocketConfig.CheckOrigin(r.Header.Get("Origin"))
	}

	// 升级HTTP连接到WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if e.config.LogConfig.Debug {
			log.Printf("[ERROR] Upgrade error: %v", err)
		}
		return
	}

	// 创建新的连接
	conn := &Connection{
		ws:         ws,
		send:       make(chan []byte, e.config.ConnectionConfig.SendChannelSize),
		engine:     e,
		closed:     false,
		closeChan:  make(chan struct{}),
		lastActive: time.Now(),
	}

	// 设置连接超时
	ws.SetReadDeadline(time.Now().Add(e.config.ConnectionConfig.ConnectionTimeout))
	ws.SetPongHandler(func(string) error {
		conn.lastActive = time.Now()
		ws.SetReadDeadline(time.Now().Add(e.config.ConnectionConfig.ConnectionTimeout))
		return nil
	})

	// 注册连接
	e.register <- conn

	// 启动读写协程
	go conn.writePump()
	go conn.readPump()
}

// readPump 处理从WebSocket读取的消息
func (c *Connection) readPump() {
	defer func() {
		c.close()
	}()

	for {
		select {
		case <-c.closeChan:
			return
		default:
			// 读取消息
			_, message, err := c.ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					if c.engine.config.LogConfig.Debug {
						log.Printf("[ERROR] Read error: %v", err)
					}
				}
				return
			}

			// 更新最后活动时间
			c.lastActive = time.Now()

			// 处理消息
			go handleMessage(message, c, c.engine)
		}
	}
}

// writePump 处理发送到WebSocket的消息
func (c *Connection) writePump() {
	ticker := time.NewTicker(c.engine.config.ConnectionConfig.HeartbeatInterval)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// 设置写入超时
			c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 通道已关闭
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 发送消息
			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				if c.engine.config.LogConfig.Debug {
					log.Printf("[ERROR] Write error: %v", err)
				}
				return
			}
		case <-ticker.C:
			// 发送心跳
			c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			// 检查连接是否超时
			if time.Since(c.lastActive) > c.engine.config.ConnectionConfig.HeartbeatTimeout {
				if c.engine.config.LogConfig.Debug {
					log.Println("[INFO] Connection timeout")
				}
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

// close 关闭连接
func (c *Connection) close() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.closeChan)
		c.engine.unregister <- c
		c.ws.Close()
	}
}
