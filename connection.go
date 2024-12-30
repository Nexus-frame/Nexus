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
		return true // 根据需要调整跨域策略
	},
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	conn := &Connection{ws: ws, send: make(chan []byte, 256)}
	hub.register <- conn

	go conn.writePump()
	go conn.readPump(hub)
}

func (c *Connection) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.ws.Close()
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
		handleMessage(message, c)
	}
}

func (c *Connection) writePump() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
	c.ws.Close()
}
