package mux

import (
	"math"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
)

const (
	HandlerLimit = math.MaxInt8/2
)

//每个请求的处理函数的接口
type RequestHandlerFunc func(ctx *Context)
type tRequestHandlerFunc = func(ctx *Context)

type handleChain []RequestHandlerFunc

type IRoute interface {
	Any(string, ...RequestHandlerFunc) IRoute
	GET(string, ...RequestHandlerFunc) IRoute
	POST(string, ...RequestHandlerFunc) IRoute
	DELETE(string, ...RequestHandlerFunc) IRoute
	PATCH(string, ...RequestHandlerFunc) IRoute
	PUT(string, ...RequestHandlerFunc) IRoute
	OPTIONS(string, ...RequestHandlerFunc) IRoute
	HEAD(string, ...RequestHandlerFunc) IRoute
	Auto(string,interface{}) IRoute

	Group(string,...RequestHandlerFunc) IRoute

	Use(...RequestHandlerFunc) IRoute

	StaticFile(string, string) IRoute
	Static(string, string) IRoute
	StaticFS(string, http.FileSystem) IRoute
}

var defaultRouterGroup = &RouterGroup{
	basePath:"/",
	handlers: nil,
	root:true,
}

//分组路由器，实现路由注册，中间件注册
type RouterGroup struct {
	mux *Mux
	basePath string
	handlers handleChain
	root bool
}

func (r *RouterGroup) GET(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodGet,relativePath,handlers)
}

func (r *RouterGroup) POST(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodPost,relativePath,handlers)
}

func (r *RouterGroup) DELETE(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodDelete,relativePath,handlers)
}

func (r *RouterGroup) PUT(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodPut,relativePath,handlers)
}

func (r *RouterGroup) PATCH(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodPatch,relativePath,handlers)
}

func (r *RouterGroup) OPTIONS(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodOptions,relativePath,handlers)
}

func (r *RouterGroup) HEAD(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	return r.handle(http.MethodHead,relativePath,handlers)
}

func (r *RouterGroup) Any(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	//嗯，简单粗暴
	r.handle(http.MethodGet,relativePath,handlers)
	r.handle(http.MethodPost,relativePath,handlers)
	r.handle(http.MethodDelete,relativePath,handlers)
	r.handle(http.MethodPut,relativePath,handlers)
	r.handle(http.MethodOptions,relativePath,handlers)
	r.handle(http.MethodPatch,relativePath,handlers)
	return	r.handle(http.MethodHead,relativePath,handlers)
}

func (r *RouterGroup) Auto(relativePath string, pkg interface{}) IRoute {
	vpkg := reflect.ValueOf(pkg)
	if !(vpkg.Kind() == reflect.Ptr && reflect.Indirect(vpkg).Kind() == reflect.Struct) {
		panic("must be a struct pointer")
	}

	for i:=0;i<vpkg.NumMethod();i++{
		spath := vpkg.Type().Method(i).Name
		path := path.Join(relativePath,spath)
		r.Any(path,vpkg.Interface().(tRequestHandlerFunc))
	}

	return r.returnObj()
}

func (r *RouterGroup) Group(relativePath string,handlers ...RequestHandlerFunc) IRoute {
	router := &RouterGroup{
		basePath:r.mergeAbsolutePath(relativePath),
		handlers:r.mergeHandlers(handlers),
		root:false,
	}
	return router
}

func (r *RouterGroup) Use(handles ...RequestHandlerFunc) IRoute {
	r.handlers = r.mergeHandlers(handles)
	return r.returnObj()
}

func (r *RouterGroup) StaticFile(relative,path string) IRoute {
	handle := func(c *Context) {
		c.File(path)
	}
	r.GET(relative, handle)
	r.HEAD(relative,handle)
	return r.returnObj()
}

func (r *RouterGroup) Static(relative string,path string) IRoute {
	r.StaticFS(relative,http.Dir(path))
	return r.returnObj()
}

func (r *RouterGroup) StaticFS(relative string,fs http.FileSystem) IRoute {
	path := filepath.Join(relative,"*filepath")

	handle := createStaticFileHandle(path,fs)

	r.GET(path,handle)
	r.HEAD(path,handle)

	return r.returnObj()
}

func createStaticFileHandle(path string,fs http.FileSystem) RequestHandlerFunc {
	handler := http.StripPrefix(path, http.FileServer(fs))
	return func(ctx *Context) {
		handler.ServeHTTP(ctx.RW,ctx.Request)
	}
}

func (r *RouterGroup) handle (method,relativePath string,handles handleChain) IRoute {
	p := r.mergeAbsolutePath(relativePath)
	chain := r.mergeHandlers(handles)
	r.mux.addRouter(method,p,chain)
	return r.returnObj()
}

//组合use注册的handle，和传入的handle
func (r *RouterGroup)mergeHandlers(chain handleChain) handleChain {
	size := len(r.handlers) + len(chain)
	if  size > int(HandlerLimit) {
		panic("too much handler, limit 63")
	}
	mergedHandlers := make([]RequestHandlerFunc,size)
	copy(mergedHandlers,r.handlers)
	copy(mergedHandlers,chain)
	return mergedHandlers
}

//合并两个路径
func (r *RouterGroup)mergeAbsolutePath(relative string) string {
	return path.Join(r.basePath,relative)
}

func (r *RouterGroup) returnObj() IRoute {
	if r.root {
		return r.mux
	}
	return r
}


