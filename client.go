package Nexus

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	// 请求超时时间
	RequestTimeout time.Duration
	// 自动重连
	AutoReconnect bool
	// 重连间隔
	ReconnectInterval time.Duration
	// 最大重连次数
	MaxReconnectAttempts int
	// 调试日志
	Debug bool
}

// DefaultClientConfig 返回默认客户端配置
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		RequestTimeout:       10 * time.Second,
		AutoReconnect:        true,
		ReconnectInterval:    5 * time.Second,
		MaxReconnectAttempts: 5,
		Debug:                false,
	}
}

// Client 结构体表示一个 Nexus 客户端
type Client struct {
	conn            *websocket.Conn
	mu              sync.Mutex
	pending         map[string]chan ResMessage
	closeChan       chan struct{}
	reconnectChan   chan struct{}
	scheme          string
	host            string
	path            string
	config          ClientConfig
	connected       bool
	reconnectCount  int
	subscriptions   map[string]HandlerFunc
	subscriptionsMu sync.Mutex
}

// NewClient 函数用于创建并连接到 Nexus 服务
func NewClient(scheme, host, path string) (*Client, error) {
	return NewClientWithConfig(scheme, host, path, DefaultClientConfig())
}

// NewClientWithConfig 使用指定配置创建并连接到 Nexus 服务
func NewClientWithConfig(scheme, host, path string, config ClientConfig) (*Client, error) {
	client := &Client{
		scheme:        scheme,
		host:          host,
		path:          path,
		pending:       make(map[string]chan ResMessage),
		closeChan:     make(chan struct{}),
		reconnectChan: make(chan struct{}),
		config:        config,
		subscriptions: make(map[string]HandlerFunc),
	}

	// 连接到服务器
	if err := client.connect(); err != nil {
		return nil, err
	}

	// 启动消息读取协程
	go client.readMessages()

	// 启动重连监控协程
	if client.config.AutoReconnect {
		go client.reconnectMonitor()
	}

	return client, nil
}

// connect 连接到WebSocket服务器
func (c *Client) connect() error {
	u := url.URL{Scheme: c.scheme, Host: c.host, Path: c.path}

	if c.config.Debug {
		log.Printf("[DEBUG] Connecting to %s", u.String())
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.reconnectCount = 0
	c.mu.Unlock()

	if c.config.Debug {
		log.Printf("[INFO] Connected to %s", u.String())
	}

	return nil
}

// reconnectMonitor 监控连接状态并在需要时重新连接
func (c *Client) reconnectMonitor() {
	for {
		select {
		case <-c.closeChan:
			return
		case <-c.reconnectChan:
			c.mu.Lock()
			if c.connected || c.reconnectCount >= c.config.MaxReconnectAttempts {
				c.mu.Unlock()
				continue
			}
			c.reconnectCount++
			attemptCount := c.reconnectCount
			c.mu.Unlock()

			if c.config.Debug {
				log.Printf("[INFO] Reconnect attempt %d/%d", attemptCount, c.config.MaxReconnectAttempts)
			}

			// 等待重连间隔
			time.Sleep(c.config.ReconnectInterval)

			// 尝试重新连接
			if err := c.connect(); err != nil {
				if c.config.Debug {
					log.Printf("[ERROR] Reconnect failed: %v", err)
				}
				// 触发下一次重连尝试
				c.reconnectChan <- struct{}{}
			} else {
				// 连接成功，重新启动消息读取
				go c.readMessages()
			}
		}
	}
}

// Req 创建请求消息
func (c *Client) Req(method string, path string, body N) *ReqMessage {
	return NewRequest(method, path, body)
}

// Do 发送请求并等待响应
func (c *Client) Do(req *ReqMessage) (*ResMessage, error) {
	resp, err := c.SendRequest(*req)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendRequest 方法用于发送请求并等待响应
func (c *Client) SendRequest(data ReqMessage) (ResMessage, error) {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return ResMessage{}, errors.New("客户端未连接")
	}

	// 确保请求有ID
	if data.ID == "" {
		data.ID = GenerateUniqueString()
	}

	// 设置时间戳
	data.Timestamp = time.Now()

	// 创建响应通道
	respChan := make(chan ResMessage, 1)
	c.pending[data.ID] = respChan

	// 发送请求
	err := c.conn.WriteMessage(websocket.TextMessage, data.Bytes())
	c.mu.Unlock()

	if err != nil {
		c.mu.Lock()
		delete(c.pending, data.ID)
		c.mu.Unlock()

		// 触发重连
		if c.config.AutoReconnect {
			c.mu.Lock()
			if c.connected {
				c.connected = false
				c.mu.Unlock()
				c.reconnectChan <- struct{}{}
			} else {
				c.mu.Unlock()
			}
		}

		return ResMessage{}, fmt.Errorf("发送请求失败: %w", err)
	}

	// 等待响应或超时
	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(c.config.RequestTimeout):
		c.mu.Lock()
		delete(c.pending, data.ID)
		c.mu.Unlock()
		return ResMessage{}, errors.New("请求超时")
	case <-c.closeChan:
		return ResMessage{}, errors.New("客户端已关闭")
	}
}

// readMessages 方法在单独的 goroutine 中运行，持续读取服务器消息
func (c *Client) readMessages() {
	defer func() {
		c.mu.Lock()
		wasConnected := c.connected
		c.connected = false
		conn := c.conn
		c.mu.Unlock()

		// 关闭连接
		if conn != nil {
			conn.Close()
		}

		// 触发重连
		if wasConnected && c.config.AutoReconnect {
			c.reconnectChan <- struct{}{}
		}
	}()

	for {
		select {
		case <-c.closeChan:
			return
		default:
			// 检查连接状态
			c.mu.Lock()
			if !c.connected {
				c.mu.Unlock()
				return
			}
			conn := c.conn
			c.mu.Unlock()

			// 读取消息
			_, message, err := conn.ReadMessage()
			if err != nil {
				if c.config.Debug {
					log.Printf("[ERROR] Read error: %v", err)
				}
				return
			}

			// 解析响应
			var resp ResMessage
			if err = json.Unmarshal(message, &resp); err != nil {
				if c.config.Debug {
					log.Printf("[ERROR] Failed to parse response: %v", err)
				}
				continue
			}

			// 处理响应
			c.mu.Lock()
			if ch, ok := c.pending[resp.ID]; ok {
				// 将响应发送到等待通道
				select {
				case ch <- resp:
				default:
					// 通道已满或已关闭，忽略响应
				}
				delete(c.pending, resp.ID)
				c.mu.Unlock()
			} else {
				c.mu.Unlock()

				// 处理订阅消息
				c.handleSubscription(resp)
			}
		}
	}
}

// Subscribe 订阅特定路径的消息
func (c *Client) Subscribe(path string, handler HandlerFunc) {
	c.subscriptionsMu.Lock()
	defer c.subscriptionsMu.Unlock()
	c.subscriptions[path] = handler
}

// Unsubscribe 取消订阅
func (c *Client) Unsubscribe(path string) {
	c.subscriptionsMu.Lock()
	defer c.subscriptionsMu.Unlock()
	delete(c.subscriptions, path)
}

// handleSubscription 处理订阅消息
func (c *Client) handleSubscription(resp ResMessage) {
	c.subscriptionsMu.Lock()
	defer c.subscriptionsMu.Unlock()

	for path, handler := range c.subscriptions {
		// TODO: 根据实际情况实现路径匹配逻辑
		if resp.Header != nil && resp.Header["path"] == path {
			// 创建上下文
			ctx := &Context{
				Response: &resp,
				Header:   resp.Header,
			}

			// 调用处理函数
			go handler(ctx)
			break
		}
	}
}

// Close 方法用于关闭与 Nexus 服务的连接
func (c *Client) Close() error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return nil
	}
	c.connected = false
	conn := c.conn
	c.mu.Unlock()

	// 关闭所有通道
	close(c.closeChan)

	// 关闭所有挂起的请求
	c.mu.Lock()
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.mu.Unlock()

	// 关闭WebSocket连接
	if conn != nil {
		return conn.Close()
	}

	return nil
}
