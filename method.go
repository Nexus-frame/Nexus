package Nexus

type method string

const (
	GET     method = "GET"
	POST    method = "POST"
	PUT     method = "PUT"
	DELETE  method = "DELETE"
	HEAD    method = "HEAD"
	OPTIONS method = "OPTIONS"
)

func allowMethod(method method) bool {
	switch method {
	case GET:
		return true
	case POST:
		return true
	case PUT:
		return true
	case DELETE:
		return true
	case HEAD:
		return false // 不放行这个请求 因为不需要
	case OPTIONS:
		return false //不放行这个请求 因为不需要
	default:
		return false
	}
}
