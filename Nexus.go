package Nexus

import "net/http"

var (
	Cors = func(r *http.Request) bool {
		return true
	}
	defaultHandler     = DefaultHandler404Handler
	defaultHandlerFlag = false
	default404Handler  = DefaultHandler404Handler
	default405Handler  = DefaultHandler405Handler
)
