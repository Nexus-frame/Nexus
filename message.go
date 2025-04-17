package Nexus

import (
	"encoding/json"
	"time"
)

// header 表示请求/响应头信息
type header map[string]any

// N 是map[string]any的别名，用于JSON对象
type N map[string]any

// status 表示状态码
type status int

// 状态码常量
const (
	StatusOK                  status = 200
	StatusCreated             status = 201
	StatusAccepted            status = 202
	StatusNoContent           status = 204
	StatusBadRequest          status = 400
	StatusUnauthorized        status = 401
	StatusForbidden           status = 403
	StatusNotFound            status = 404
	StatusMethodNotAllowed    status = 405
	StatusConflict            status = 409
	StatusInternalServerError status = 500
	StatusNotImplemented      status = 501
	StatusBadGateway          status = 502
	StatusServiceUnavailable  status = 503
)

// ReqMessage 表示客户端发送的请求消息
type ReqMessage struct {
	ID        string    `json:"id,omitempty"`        // 唯一请求ID，用于响应匹配
	Path      string    `json:"path,omitempty"`      // 请求路径
	Method    string    `json:"method,omitempty"`    // 请求类型，如 "GET", "POST" 等
	Params    Params    `json:"params,omitempty"`    // 路由参数
	Header    header    `json:"header,omitempty"`    // 请求头
	Body      any       `json:"body,omitempty"`      // 请求参数
	Timestamp time.Time `json:"timestamp,omitempty"` // 请求时间戳
}

// ResMessage 表示服务端发送的响应消息
type ResMessage struct {
	ID        string    `json:"id,omitempty"`        // 唯一请求ID，用于响应匹配
	Status    status    `json:"status,omitempty"`    // 状态码
	Header    header    `json:"header,omitempty"`    // 响应头
	Body      any       `json:"body,omitempty"`      // 响应参数
	Timestamp time.Time `json:"timestamp,omitempty"` // 响应时间戳
}

// 默认消息模板
var DefaultReqMessage = ReqMessage{
	ID:        "",
	Path:      "",
	Method:    "",
	Header:    make(header),
	Body:      nil,
	Timestamp: time.Time{},
}

var DefaultResMessage = ResMessage{
	ID:        "",
	Status:    StatusOK,
	Header:    make(header),
	Body:      nil,
	Timestamp: time.Time{},
}

// 默认请求头
var DefaultHeader = header{
	"Content-Type": "application/json",
	"Accept":       "application/json",
	"User-Agent":   "Nexus",
}

// Bytes 将ReqMessage序列化为JSON字节数组
func (r *ReqMessage) Bytes() []byte {
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}

	data, err := json.Marshal(r)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// Bytes 将ResMessage序列化为JSON字节数组
func (r *ResMessage) Bytes() []byte {
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}

	data, err := json.Marshal(r)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// NewRequest 创建一个新的请求消息
func NewRequest(method, path string, body any) *ReqMessage {
	return &ReqMessage{
		ID:        GenerateUniqueString(),
		Method:    method,
		Path:      path,
		Header:    DefaultHeader,
		Body:      body,
		Timestamp: time.Now(),
	}
}

// NewResponse 创建一个新的响应消息
func NewResponse(id string, status status, body any) *ResMessage {
	return &ResMessage{
		ID:        id,
		Status:    status,
		Header:    DefaultHeader,
		Body:      body,
		Timestamp: time.Now(),
	}
}
