package mux

import (
	"net/http"
)

//http上下文信息
// 1.参数解析
// 2.格式化返回值功能
// 3.路由传值
type Context struct {
	Request *http.Request
	RW http.ResponseWriter
}