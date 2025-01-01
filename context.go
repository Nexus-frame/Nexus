package Nexus

type Context struct {
	Request    *ReqMessage
	Response   *ResMessage
	Header     header
	connection *Connection
	Keys       map[string]any
	Errors     []*error
	exit       bool
}

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
	}
}

func (c *Context) JSON(Status status, data N) {
	c.Response = &ResMessage{
		Header: c.Header,
		ID:     c.Request.ID,
		Status: Status,
		Body:   data,
	}
	c.Exit()
}

func (c *Context) Get(key string) any {
	return c.Keys[key]
}

func (c *Context) Set(key string, value any) {
	c.Keys[key] = value
}

func (c *Context) Delete(key string) {
	delete(c.Keys, key)
}

func (c *Context) GetHeader(key string) any {
	return c.Header[key]
}

func (c *Context) SetHeader(key string, value any) {
	c.Header[key] = value
}

func (c *Context) DeleteHeader(key string) {
	delete(c.Header, key)
}

func (c *Context) Exit() {
	c.exit = true
}
