package Nexus

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Engine struct {
	Port          int    // 端口 如果端口为0 则表示使用其他的http服务来集成nexus
	WebSocketPath string // ws 路径

	connections map[*Connection]bool
	broadcast   chan []byte
	register    chan *Connection
	unregister  chan *Connection
	mu          sync.Mutex
}

type Context struct {
	Request    *ReqMessage
	Response   *ResMessage
	Header     header
	connection *Connection
	Keys       map[string]any
	Errors     []*error
}

func New() *Engine {
	return &Engine{
		Port:          8080,
		WebSocketPath: "/",
	}
}

func (c *Context) JSON(Status status, data N) {
	c.Response = &ResMessage{
		Header: c.Header,
		ID:     c.Request.ID,
		Status: Status,
		Body:   data,
	}
	respBytes, _ := json.Marshal(c.Response)
	c.connection.send <- respBytes
}

func (e *Engine) WebSocketService() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serveWs(e, w, r)
	}
}
func (e *Engine) Run(port int) {
	go e.run() //启动ws连接管理服务
	//serveWs()
}

func (e *Engine) run() {
	for {
		select {
		case conn := <-e.register:
			e.mu.Lock()
			e.connections[conn] = true
			e.mu.Unlock()
		case conn := <-e.unregister:
			e.mu.Lock()
			if _, ok := e.connections[conn]; ok {
				delete(e.connections, conn)
				close(conn.send)
			}
			e.mu.Unlock()
		case message := <-e.broadcast:
			e.mu.Lock()
			for conn := range e.connections {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(e.connections, conn)
				}
			}
			e.mu.Unlock()
		}
	}
}
