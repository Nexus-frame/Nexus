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

func Check(method string) bool {
	switch method {
	case "GET":
		return true
	case "POST":
		return true
	case "PUT":
		return true
	case "DELETE":
		return true
	case "HEAD":
		return true
	case "OPTIONS":
		return true
	default:
		return false
	}
}
