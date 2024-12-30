package Nexus

type path string
type header map[string]any
type payload map[string]any

type ReqMessage struct {
	Path    path    `json:"path"`    // 请求路径
	Method  method  `json:"method"`  // 请求类型，如 "GET", "POST" 等
	ID      string  `json:"id"`      // 唯一请求 ID，用于响应匹配
	Header  header  `json:"header"`  // 请求头
	Payload payload `json:"payload"` // 请求参数
}

type ResMessage struct {
	ID      string  `json:"id"`      // 唯一请求 ID，用于响应匹配
	Header  header  `json:"header"`  // 响应头
	Payload payload `json:"payload"` // 响应参数
}
