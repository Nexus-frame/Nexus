package Nexus

type Context struct {
	Request    *ReqMessage
	Response   *ResMessage
	Header     header
	connection *Connection
	Keys       map[string]any
	Errors     []*error
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
	}
}
