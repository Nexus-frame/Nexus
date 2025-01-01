package Nexus

import (
	"encoding/json"
	"fmt"
	"log"
)

type HandlerFunc func(c *Context)

type HandlerFuncList []HandlerFunc

var DefaultHandlerFuncList = HandlerFuncList{DefaultHandler404Handler}

func handleMessage(message []byte, conn *Connection, e *Engine) {
	var c = NewContext(conn)
	var handler HandlerFuncList = nil
	if err := json.Unmarshal(message, &c.Request); err != nil {
		handler = DefaultHandlerFuncList
	}
	if handler == nil {

		h, params, ok := e.ParsePath(c.Request.Method, c.Request.Path)
		if !ok {
			// 未找到对应的处理器 调用默认处理器
			DefaultHandler404Handler(c)
		} else {
			c.Request.Params = params
			// 调用对应的处理器
			for _, handlerFunc := range h {
				if !c.exit {
					handlerFunc(c)
				}
			}
		}

	} else {
		DefaultHandler500Handler(c)
	}
	respBytes, err := json.Marshal(c.Response)
	if err != nil {
		log.Println("Error marshalling createOrder response:", err)

		respBytes = []byte(fmt.Sprintf(`{"id":"%s","header":{"Content-Type":"application/json"},"body":{"error":"404 Not Found"}}`, c.Request.ID))
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
