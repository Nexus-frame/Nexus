package Nexus

import "log"

// Context 表示一个请求上下文
type Context struct {
	Request    *ReqMessage
	Response   *ResMessage
	Header     header
	connection *Connection
	Keys       map[string]any
	Errors     []*error
	exit       bool
	index      int8
	handlers   HandlerFuncList
}

// NewContext 创建一个新的上下文
func NewContext(conn *Connection) *Context {
	var defaultReqMessage = DefaultReqMessage
	var defaultResMessage = DefaultResMessage
	return &Context{
		Request:    &defaultReqMessage,
		Response:   &defaultResMessage,
		connection: conn,
		Keys:       make(map[string]any),
		Errors:     make([]*error, 0),
		exit:       false,
		index:      -1,
		handlers:   nil,
	}
}

// Next 调用下一个处理器
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// IsAborted 检查处理是否已中止
func (c *Context) IsAborted() bool {
	return c.exit
}

// JSON 设置JSON响应
func (c *Context) JSON(Status status, data N) {
	c.Response = &ResMessage{
		Header: c.Header,
		ID:     c.Request.ID,
		Status: Status,
		Body:   data,
	}
	c.Exit()
}

// Get 获取上下文中的键值
func (c *Context) Get(key string) any {
	return c.Keys[key]
}

// Set 设置上下文中的键值
func (c *Context) Set(key string, value any) {
	c.Keys[key] = value
}

// Delete 删除上下文中的键值
func (c *Context) Delete(key string) {
	delete(c.Keys, key)
}

// GetHeader 获取请求头
func (c *Context) GetHeader(key string) any {
	return c.Header[key]
}

// SetHeader 设置请求头
func (c *Context) SetHeader(key string, value any) {
	c.Header[key] = value
}

// AddHeader 添加请求头（如果不存在）
func (c *Context) AddHeader(key string, value any) {
	if _, ok := c.Header[key]; !ok {
		c.Header[key] = value
	}
}

// AddHeaders 批量添加请求头（如果不存在）
func (c *Context) AddHeaders(h header) {
	for k, v := range h {
		if _, ok := c.Header[k]; !ok {
			c.Header[k] = v
		}
	}
}

// SetHeaders 批量设置请求头
func (c *Context) SetHeaders(h map[string]any) {
	for k, v := range h {
		c.Header[k] = v
	}
}

// DeleteHeader 删除请求头
func (c *Context) DeleteHeader(key string) {
	delete(c.Header, key)
}

// Exit 中止处理链
func (c *Context) Exit() {
	c.exit = true
	c.index = abortIndex
}

// Send 发送响应
func (c *Context) Send(data []byte) {
	if c.connection != nil {
		select {
		case c.connection.send <- data:
			// 已发送
		default:
			// 如果通道已满，记录错误
			if c.connection.engine.config.LogConfig.Debug {
				log.Printf("[ERROR] Send channel full, message discarded")
			}
		}
	}
}

// Error 添加错误到上下文
func (c *Context) Error(err error) {
	if err != nil {
		c.Errors = append(c.Errors, &err)
	}
}
