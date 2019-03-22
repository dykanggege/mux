package mux

import (
	"net/http"
	"sync"
)

//使用默认路由
func Default() *Mux {
	return New()
}

//使用一个新建的路由
func New() *Mux {
	return &Mux{
		ctxPool:&sync.Pool{New: func() interface{} {
			return &Context{}
		}},
	}
}

//使用配置文件创建一个新的路由
func NewConf(conf ...string) *Mux {
	return &Mux{}
}

type Mux struct {
	RouterGroup
	ctxPool *sync.Pool
	trees methodTrees
}

func (m *Mux) ServeHTTP(rw http.ResponseWriter,req *http.Request) {
	//从池子中取出 context，http上下文信息
	ctx := m.ctxPool.Get().(*Context)
	ctx.Request = req
	ctx.RW = rw

	m.handleHTTPRequest(ctx)

	m.ctxPool.Put(ctx)
}

func (m *Mux)handleHTTPRequest(ctx *Context)  {
	method := ctx.Request.Method
	//path := ctx.Request.URL.Path

	trees := m.trees
	for i := range trees{
		if method != trees[i].method{
			continue
		}
		//t = trees[i]
	}
}

func (m *Mux)addRouter(method,absolutePath string,chain handleChain)  {

}