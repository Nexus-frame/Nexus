package Nexus

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Connection struct {
	ws   *websocket.Conn
	send chan []byte
}

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsHandler func(e *Engine, w http.ResponseWriter, r *http.Request)

func serveWs(e *Engine, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	conn := &Connection{ws: ws, send: make(chan []byte, 256)}
	e.register <- conn // 注册连接
	go conn.writePump()
	go conn.readPump(e)
}

func (c *Connection) readPump(e *Engine) {
	defer func() {
		e.unregister <- c // 注销连接
		err := c.ws.Close()
		if err != nil {
			log.Println("Close error:", err)
		}
	}()
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %v", err)
			}
			break
		}
		// 处理接收到的消息
		handleMessage(message, c, e)
	}
}

func (c *Connection) writePump() {
	defer func() {
		err := c.ws.Close()
		if err != nil {
			log.Println("Close error:", err)
		}
	}()
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
