package Nexus

import (
	"math"
	"net/http"
	"sync"
)

const abortIndex int8 = math.MaxInt8 >> 1

type Engine struct {
	RouterGroup
	Port          int    // 端口 如果端口为0 则表示使用其他的http服务来集成nexus
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
		Port:          0,
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

func (c *Context) JSON(Status status, data N) {
	c.Response = &ResMessage{
		Header: c.Header,
		ID:     c.Request.ID,
		Status: Status,
		Body:   data,
	}
}

func (e *Engine) WebSocketService() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serveWs(e, w, r)
	}
}
func (e *Engine) Run() {
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
