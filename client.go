package Nexus

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"net/url"
	"sync"
	"time"
)

// Client 结构体表示一个 Nexus 客户端
type Client struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	pending   map[string]chan ResMessage
	closeChan chan struct{}
}

func (c *Client) Req(method string, path string, body N) (data ReqMessage) {
	data.Method = method
	data.Path = path
	data.Body = body
	data.Header = DefaultHeader
	return
}

func (c *Client) Do(req ReqMessage) (data ResMessage, err error) {
	data, err = c.SendRequest(req)
	if err != nil {
		return
	}
	return
}

// NewClient 函数用于创建并连接到 Nexus 服务
func NewClient(scheme, host, path string) (*Client, error) {
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:      conn,
		pending:   make(map[string]chan ResMessage),
		closeChan: make(chan struct{}),
	}
	go client.readMessages()
	return client, nil
}

// SendRequest 方法用于发送请求并等待响应
func (c *Client) SendRequest(data ReqMessage) (ResMessage, error) {
	data.ID = UUID()
	respChan := make(chan ResMessage)
	c.mu.Lock()
	c.pending[data.ID] = respChan
	err := c.conn.WriteMessage(websocket.TextMessage, data.Bytes())
	if err != nil {
		return ResMessage{}, err
	}
	c.mu.Unlock()
	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(10 * time.Second):
		return ResMessage{}, errors.New("请求超时")
	}
}

// readMessages 方法在单独的 goroutine 中运行，持续读取服务器消息
func (c *Client) readMessages() {
	for {
		select {
		case <-c.closeChan:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				// 处理读取错误，例如记录日志或重新连接
				return
			}
			var resp ResMessage
			if err = json.Unmarshal(message, &resp); err != nil {
				// 处理 JSON 解析错误
				continue
			}
			c.mu.Lock()
			if ch, ok := c.pending[resp.ID]; ok {
				ch <- resp
				close(ch)
				delete(c.pending, resp.ID)
			}
			c.mu.Unlock()
		}
	}
}

// Close 方法用于关闭与 Nexus 服务的连接
func (c *Client) Close() error {
	close(c.closeChan)
	return c.conn.Close()
}
