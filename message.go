package Nexus

type path string
type header map[string]any
type N map[string]any
type status int
type ReqMessage struct {
	ID     string `json:"id"`     // 唯一请求 ID，用于响应匹配
	Path   path   `json:"path"`   // 请求路径
	Method method `json:"method"` // 请求类型，如 "GET", "POST" 等
	Header header `json:"header"` // 请求头
	Body   N      `json:"body"`   // 请求参数
}

type ResMessage struct {
	ID     string `json:"id"`     // 唯一请求 ID，用于响应匹配
	Status status `json:"status"` //状态码
	Header header `json:"header"` // 响应头
	Body   N      `json:"body"`   // 响应参数
}
