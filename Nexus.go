package Nexus

import (
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"net/http"
	"sync"
)

const abortIndex int8 = math.MaxInt8 >> 1

type Engine struct {
	RouterGroup
	Addr          string // 端口 如果端口为0 则表示使用其他的http服务来集成nexus
	WebSocketPath string // ws 路径
	connections   map[*Connection]bool
	broadcast     chan []byte
	register      chan *Connection
	unregister    chan *Connection
	mu            sync.Mutex
	trees         methodTrees
}

var _ IRouter = (*Engine)(nil)

func New() *Engine {
	e := &Engine{
		Addr:          ":8080",
		WebSocketPath: "/",
		trees:         make(methodTrees, 0, 9),
		connections:   make(map[*Connection]bool),
		broadcast:     make(chan []byte),
		register:      make(chan *Connection),
		unregister:    make(chan *Connection),
	}
	e.RouterGroup.engine = e
	go e.run() // 启动主循环协程
	return e
}

func (e *Engine) WebSocketService() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serveWs(e, w, r)
	}
}

func (e *Engine) GinServe(c *gin.Context) {
	//检测是否为ws请求
	if c.GetHeader("Upgrade") != "websocket" {
		var h = make(header)
		h["User-Agent"] = c.GetHeader("User-Agent")
		h["Accept-Encoding"] = c.GetHeader("Accept-Encoding")
		h["Accept-Language"] = c.GetHeader("Accept-Language")
		h["Host"] = c.GetHeader("Host")
		h["Origin"] = c.GetHeader("Origin")
		for k, v := range h {
			if v == "" {
				delete(h, k)
			}
		}
		c.JSON(200, h)
		return
	}
	serveWs(e, c.Writer, c.Request)
}

func (e *Engine) Run(addr string, path ...string) {
	wsPath := "/"
	if len(path) > 0 {
		wsPath = path[0]
	}
	http.HandleFunc(wsPath, e.WebSocketService())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
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

func (e *Engine) addRoute(method, path string, handlers HandlerFuncList) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	root := e.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		e.trees = append(e.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}
