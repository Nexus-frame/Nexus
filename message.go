package Nexus

type header map[string]any
type N map[string]any
type status int
type ReqMessage struct {
	ID     string `json:"id,omitempty"`     // 唯一请求 ID，用于响应匹配
	Path   string `json:"path,omitempty"`   // 请求路径
	Method string `json:"method,omitempty"` // 请求类型，如 "GET", "POST" 等
	Params Params `json:"params,omitempty"`
	Header header `json:"header,omitempty"` // 请求头
	Body   N      `json:"body,omitempty"`   // 请求参数
}

type ResMessage struct {
	ID     string `json:"id,omitempty"`     // 唯一请求 ID，用于响应匹配
	Status status `json:"status,omitempty"` //状态码
	Header header `json:"header,omitempty"` // 响应头
	Body   N      `json:"body,omitempty"`   // 响应参数
}

var DefaultReqMessage = ReqMessage{
	ID:     "",
	Path:   "",
	Method: "",
	Header: nil,
	Body:   nil,
}

var DefaultResMessage = ResMessage{
	ID:     "",
	Status: 0,
	Header: nil,
	Body:   nil,
}
