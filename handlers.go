package Nexus

import (
	"encoding/json"
	"log"
)

type HandlerFunc func(c *Context)

type HandlerFuncList []HandlerFunc

var DefaultHandlerFuncList = HandlerFuncList{DefaultHandler404Handler}

var handlers = map[method]map[path]HandlerFuncList{}

func handleMessage(message []byte, conn *Connection) {
	var c = NewContext(conn)
	var handler HandlerFuncList = nil
	if err := json.Unmarshal(message, &c.Request); err != nil {
		handler = DefaultHandlerFuncList
	}
	if handler == nil {
		h, exists := handlers[c.Request.Method][c.Request.Path]
		if !exists {
			// 未找到对应的处理器 调用默认处理器
			DefaultHandler404Handler(c)
		}
		// 调用对应的处理器
		for _, handlerFunc := range h {
			handlerFunc(c)
		}
	} else {
		DefaultHandler500Handler(c)
	}
	respBytes, err := json.Marshal(c.Response)
	if err != nil {
		log.Println("Error marshalling createOrder response:", err)
		return
	}
	conn.send <- respBytes
}

// DefaultHandler404Handler 默认404处理器
func DefaultHandler404Handler(c *Context) {
	c.Response = &ResMessage{
		Header: header{
			"Content-Type": "application/json",
		},
		ID: c.Request.ID,
		Body: N{
			"error": "404 Not Found",
		},
	}
}

// DefaultHandler500Handler 默认404处理器
func DefaultHandler500Handler(c *Context) {
	c.Response = &ResMessage{
		Header: header{
			"Content-Type": "application/json",
		},
		ID: c.Request.ID,
		Body: N{
			"error": "500 Internal Server Error",
		},
	}
}
