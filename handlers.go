package Nexus

import (
	"encoding/json"
	"log"
)

type HandlerFunc func(payload payload, conn *Connection)

var handlers = map[method]map[path]HandlerFunc{}

func handleMessage(message []byte, conn *Connection) {
	var req ReqMessage
	if err := json.Unmarshal(message, &req); err != nil {
		log.Println("Invalid message format:", err)
		return
	}

	handler, exists := handlers[req.Method][req.Path]
	if !exists {
		// 未找到对应的处理器 调用默认处理器
		handler = DefaultHandler404Handler
	}
	// 调用对应的处理器
	handler(req.Payload, conn)
}

// DefaultHandler404Handler 默认404处理器
func DefaultHandler404Handler(payload payload, conn *Connection) {
	response := map[string]interface{}{
		"error": "404 Not Found",
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling createOrder response:", err)
		return
	}
	conn.send <- respBytes
}

func DefaultHandler405Handler(payload json.RawMessage, conn *Connection) {
	response := map[string]interface{}{
		"error": "405 method not allowed",
	}
	respBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling createOrder response:", err)
		return
	}
	conn.send <- respBytes
}
