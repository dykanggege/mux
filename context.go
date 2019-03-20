package mux

import (
	"net/http"
	"sync"
)

var ctxPool *sync.Pool


func init() {
	ctxPool = &sync.Pool{
		New: func() interface{} {
			return &Context{}
		},
	}
}

// http请求的上下文信息
type Context struct {
	Request *http.Request
	RW http.ResponseWriter
}