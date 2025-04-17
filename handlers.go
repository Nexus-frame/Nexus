package Nexus

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// HandlerFunc 是Nexus处理函数的类型
type HandlerFunc func(c *Context)

// HandlerFuncList 是HandlerFunc的切片
type HandlerFuncList []HandlerFunc

// 默认处理函数
var DefaultHandlerFuncList = HandlerFuncList{DefaultHandler404Handler}

// handleMessage 处理接收到的WebSocket消息
func handleMessage(message []byte, conn *Connection, e *Engine) {
	start := time.Now()
	var c = NewContext(conn)
	var requestID string

	// 解析请求消息
	if err := json.Unmarshal(message, &c.Request); err != nil {
		if e.config.LogConfig.Debug {
			log.Printf("[ERROR] Failed to parse message: %v", err)
		}
		// 如果无法解析消息，返回500错误
		c.Response = &ResMessage{
			Header: DefaultHeader,
			Status: StatusInternalServerError,
			Body: N{
				"error":  "Invalid message format",
				"detail": err.Error(),
			},
		}
		sendResponse(c, conn, e)
		return
	}

	requestID = c.Request.ID

	// 记录访问日志
	if e.config.LogConfig.AccessLog {
		log.Printf("[ACCESS] %s %s %s", requestID, c.Request.Method, c.Request.Path)
	}

	// 路由分发
	handlers, params, ok := e.ParsePath(c.Request.Method, c.Request.Path)
	if !ok {
		// 未找到路由，调用404处理函数
		if e.config.LogConfig.Debug {
			log.Printf("[DEBUG] Route not found: %s %s", c.Request.Method, c.Request.Path)
		}
		c.handlers = HandlerFuncList{DefaultHandler404Handler}
	} else {
		// 设置路由参数和请求头
		c.Request.Params = params
		c.Header = c.Request.Header
		c.handlers = handlers
	}

	// 执行中间件和处理函数链
	c.Next()

	// 如果没有设置响应，设置默认响应
	if c.Response == nil || c.Response.ID == "" {
		c.Response = &ResMessage{
			ID:     requestID,
			Header: c.Header,
			Status: StatusOK,
			Body:   N{},
		}
	}

	// 发送响应
	sendResponse(c, conn, e)

	// 记录处理时间
	if e.config.LogConfig.Debug {
		elapsed := time.Since(start)
		log.Printf("[DEBUG] Request %s processed in %v", requestID, elapsed)
	}
}

// sendResponse 发送响应到客户端
func sendResponse(c *Context, conn *Connection, e *Engine) {
	// 确保响应ID与请求ID一致
	if c.Response.ID == "" {
		c.Response.ID = c.Request.ID
	}

	// 设置时间戳
	if c.Response.Timestamp.IsZero() {
		c.Response.Timestamp = time.Now()
	}

	// 序列化响应
	respBytes, err := json.Marshal(c.Response)
	if err != nil {
		if e.config.LogConfig.Debug {
			log.Printf("[ERROR] Failed to serialize response: %v", err)
		}
		// 如果无法序列化响应，返回简单的错误响应
		respBytes = []byte(fmt.Sprintf(`{"id":"%s","status":%d,"header":{"Content-Type":"application/json"},"body":{"error":"Internal Server Error"}}`,
			c.Request.ID, StatusInternalServerError))
	}

	// 发送响应
	select {
	case conn.send <- respBytes:
		// 消息已发送
	default:
		// 发送通道已满，关闭连接
		if e.config.LogConfig.Debug {
			log.Printf("[WARN] Send channel full, closing connection")
		}
		conn.close()
	}
}

// DefaultHandler404Handler 默认404处理函数
func DefaultHandler404Handler(c *Context) {
	c.Response = &ResMessage{
		Header: DefaultHeader,
		ID:     c.Request.ID,
		Status: StatusNotFound,
		Body: N{
			"error":   "Not Found",
			"message": fmt.Sprintf("No route found for %s %s", c.Request.Method, c.Request.Path),
		},
	}
	c.Exit()
}

// DefaultHandler500Handler 默认500处理函数
func DefaultHandler500Handler(c *Context) {
	c.Response = &ResMessage{
		Header: DefaultHeader,
		ID:     c.Request.ID,
		Status: StatusInternalServerError,
		Body: N{
			"error": "Internal Server Error",
		},
	}
	c.Exit()
}

// DefaultHandlerMethodNotAllowedHandler 默认405处理函数
func DefaultHandlerMethodNotAllowedHandler(c *Context) {
	c.Response = &ResMessage{
		Header: DefaultHeader,
		ID:     c.Request.ID,
		Status: StatusMethodNotAllowed,
		Body: N{
			"error":   "Method Not Allowed",
			"message": fmt.Sprintf("Method %s not allowed for path %s", c.Request.Method, c.Request.Path),
		},
	}
	c.Exit()
}
