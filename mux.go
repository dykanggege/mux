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
func New() (m *Mux) {
	m = &Mux{
		RouterGroup:defaultRouterGroup,
		ctxPool:&sync.Pool{},
		trees:methodTrees{},
	}

	m.RouterGroup.mux = m
	m.ctxPool.New = func() interface{} {
		c := &Context{mux:m}
		return 	c
	}
	return
}

//使用配置文件创建一个新的路由
//func NewConf(conf ...string) *Mux {
//	return &Mux{}
//}

type Mux struct {
	RouterGroup

	//if true 使用未解码的PATH，default false
	UseRawPath bool

	//不转义的path参数
	//if useRawPath is false,一定是非转义(解码的)，如果直接使用url.path，则也是非转义的
	UnescapePathValues bool

	// Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	ctxPool *sync.Pool
	trees methodTrees
}

func (m *Mux) ServeHTTP(rw http.ResponseWriter,req *http.Request) {
	//从池子中取出 context，http上下文信息
	ctx := m.ctxPool.Get().(*Context)
	ctx.Request = req
	ctx.Writer = rw
	ctx.reset()

	m.handleHTTPRequest(ctx)

	m.ctxPool.Put(ctx)
}

func (m *Mux)handleHTTPRequest(ctx *Context)  {
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	handlers, params, ok := m.trees.getValue(method, path, nil, false)
	if handlers != nil{
		ctx.handlers = handlers
		ctx.params = params
		ctx.Next()
		return
	}
}

func (m *Mux)addRouter(method,absolutePath string,chain handleChain)  {
	m.trees.addRouter(method,absolutePath,chain)
}