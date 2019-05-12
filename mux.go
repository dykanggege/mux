package mux

import (
	"mux/ctx"
	"mux/router"
	"mux/session"
	"net/http"
	"sync"
)

func Default() *Mux {
	m := New()
	//TODO:Default注入
	//注入session
	m.Session = &session.DefaultSession{}
	//注入log
	//注入recover
	return m
}

//使用一个新建的路由
//TODO:限制连接的最大数量
//TODO:支持fasthttp
func New() (m *Mux) {
	m = &Mux{
		RouterGroup: router.defaultRouterGroup,
		ctxPool:     &sync.Pool{},
		trees:       router.methodTrees{},
	}

	m.RouterGroup.mux = m
	m.ctxPool.New = func() interface{} {
		c := &ctx.Context{mux: m}
		return 	c
	}
	return
}

//使用配置文件创建一个新的路由
//func NewConf(conf ...string) *Mux {
//	return &Mux{}
//}

type Mux struct {
	router.RouterGroup
	Session session.Sessioner

	//if true 使用未解码的PATH，default false
	UseRawPath bool

	//不转义的path参数
	//if useRawPath is false,一定是非转义(解码的)，如果直接使用url.path，则也是非转义的
	UnescapePathValues bool

	// Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	ctxPool *sync.Pool
	trees   router.methodTrees
}

func (m *Mux) ServeHTTP(rw http.ResponseWriter,req *http.Request) {
	//从池子中取出 Context，http上下文信息
	ctx := m.ctxPool.Get().(*ctx.Context)
	ctx.reset(req,rw)

	m.handleHTTPRequest(ctx)

	m.ctxPool.Put(ctx)
}

func (m *Mux)handleHTTPRequest(ctx *ctx.Context)  {
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	//TODO:解析配置文件，根据文件信息作调整
	handlers, params, _ := m.trees.getValue(method, path, nil, false)
	if handlers != nil{
		ctx.handlers = handlers
		ctx.params = params
		ctx.Next()
		return
	}
}

func (m *Mux)addRouter(method,absolutePath string,chain router.handleChain)  {
	m.trees.addRouter(method,absolutePath,chain)
}