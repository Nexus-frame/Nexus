package Nexus

import (
	"context"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const abortIndex int8 = math.MaxInt8 >> 1

// Engine 是Nexus框架的核心结构
type Engine struct {
	RouterGroup
	config       Config       // 配置
	server       *http.Server // HTTP服务器
	connections  map[*Connection]bool
	broadcast    chan []byte
	register     chan *Connection
	unregister   chan *Connection
	mu           sync.Mutex
	trees        methodTrees
	shuttingDown bool
	shutdownChan chan struct{}
}

var _ IRouter = (*Engine)(nil)

// New 创建一个新的Nexus引擎实例
func New() *Engine {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig 使用指定配置创建一个新的Nexus引擎实例
func NewWithConfig(config Config) *Engine {
	e := &Engine{
		config:       config,
		trees:        make(methodTrees, 0, 9),
		connections:  make(map[*Connection]bool),
		broadcast:    make(chan []byte),
		register:     make(chan *Connection),
		unregister:   make(chan *Connection),
		shutdownChan: make(chan struct{}),
	}
	e.RouterGroup.engine = e
	e.RouterGroup.root = true

	// 启动主循环协程
	go e.run()

	return e
}

// WebSocketService 返回处理WebSocket连接的http.HandlerFunc
func (e *Engine) WebSocketService() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if e.shuttingDown {
			http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
			return
		}
		serveWs(e, w, r)
	}
}

// Run 启动WebSocket服务器
func (e *Engine) Run(args ...string) error {
	// 处理命令行参数
	switch len(args) {
	case 1:
		e.config.WebSocketConfig.Port = args[0]
	case 2:
		e.config.WebSocketConfig.Port = args[0]
		e.config.WebSocketConfig.Path = args[1]
	}

	// 设置HTTP处理函数
	addr := ":" + e.config.WebSocketConfig.Port
	path := e.config.WebSocketConfig.Path

	// 创建HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc(path, e.WebSocketService())

	e.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 设置优雅关闭
	go e.gracefulShutdown()

	// 启动服务器
	if e.config.LogConfig.Debug {
		log.Printf("[INFO] Server starting on %s with WebSocket path %s", addr, path)
	}

	if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("[ERROR] ListenAndServe: %v", err)
		return err
	}

	// 等待优雅关闭完成
	<-e.shutdownChan
	if e.config.LogConfig.Debug {
		log.Printf("[INFO] Server stopped gracefully")
	}

	return nil
}

// run 处理连接注册、注销和广播消息
func (e *Engine) run() {
	for {
		select {
		case conn := <-e.register:
			e.mu.Lock()
			e.connections[conn] = true
			e.mu.Unlock()
			if e.config.LogConfig.Debug {
				log.Printf("[INFO] Client connected, total connections: %d", len(e.connections))
			}
		case conn := <-e.unregister:
			e.mu.Lock()
			if _, ok := e.connections[conn]; ok {
				delete(e.connections, conn)
				close(conn.send)
				if e.config.LogConfig.Debug {
					log.Printf("[INFO] Client disconnected, total connections: %d", len(e.connections))
				}
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

// Broadcast 向所有连接的客户端广播消息
func (e *Engine) Broadcast(message []byte) {
	e.broadcast <- message
}

// gracefulShutdown 处理服务器的优雅关闭
func (e *Engine) gracefulShutdown() {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if e.config.LogConfig.Debug {
		log.Printf("[INFO] Server is shutting down...")
	}

	e.shuttingDown = true

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := e.server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server shutdown error: %v", err)
	}

	// 关闭所有WebSocket连接
	e.mu.Lock()
	for conn := range e.connections {
		conn.close()
	}
	e.mu.Unlock()

	// 发送关闭完成信号
	close(e.shutdownChan)
}

// addRoute 添加路由处理函数
func (e *Engine) addRoute(method, path string, handlers HandlerFuncList) {
	if e.config.LogConfig.Debug {
		log.Printf("[DEBUG] Route added: %s %s", method, path)
	}

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
