package route

import (
	"math"
	"mux/session"
	"net/http"
	"path"
	"reflect"
	"sync"
)

const (
	//一次请求中最大handler的个数
	HandlerLimit = math.MaxInt8/2
)

//每个请求的处理函数的接口
type HandlerFunc func(ctx *Context)
type tRequestHandlerFunc = HandlerFunc

type Router interface {
	//注册路由
	ANY(string, ...HandlerFunc) Router
	GET(string, ...HandlerFunc) Router
	POST(string, ...HandlerFunc) Router
	DELETE(string, ...HandlerFunc) Router
	PATCH(string, ...HandlerFunc) Router
	PUT(string, ...HandlerFunc) Router
	OPTIONS(string, ...HandlerFunc) Router
	HEAD(string, ...HandlerFunc) Router
	AUTO(string,interface{}) Router

	//分组路由
	Group(string,...HandlerFunc) Router

	//AOP切面编程
	//TODO:更多层次的切面
	Use(...HandlerFunc) Router
}

var ctxpool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

var DefaultRoute = Route{
	RouteConf:	  &Config{
		PathUnescape: true,
		MaxMultipartMemory: 4 << 20,
		OpenSession:  true,
	},
	tree:     NewMethodTrees(),
	//manager:s
	basePath: "/",
	Handlers: nil,
}

//分组路由器，实现路由注册，中间件注册
type Route struct {
	RouteConf *Config
	tree      *MethodTrees
	manager   *session.Manager
	basePath  string
	Handlers  []HandlerFunc
}
//配置文件
type Config struct {
	//是否将path转义用于字典树匹配
	PathUnescape bool
	//request表单提交中使用的内存限制
	MaxMultipartMemory int64
	//是否启用session
	OpenSession bool
}

func New(conf *Config,manager *session.Manager) *Route {
	return &Route{
		RouteConf: conf,
		tree:      NewMethodTrees(),
		manager:   manager,
		basePath:  "/",
		Handlers:  nil,
	}
}

//注册路由
func (r *Route) GET(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodGet,relativePath,handlers)
}

func (r *Route) POST(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodPost,relativePath,handlers)
}

func (r *Route) DELETE(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodDelete,relativePath,handlers)
}

func (r *Route) PUT(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodPut,relativePath,handlers)
}

func (r *Route) PATCH(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodPatch,relativePath,handlers)
}

func (r *Route) OPTIONS(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodOptions,relativePath,handlers)
}

func (r *Route) HEAD(relativePath string,handlers ...HandlerFunc) Router {
	return r.handle(http.MethodHead,relativePath,handlers)
}
//匹配上面任意的请求方式
func (r *Route) ANY(relativePath string,handlers ...HandlerFunc) Router {
	//嗯，简单粗暴
	r.handle(http.MethodGet,relativePath,handlers)
	r.handle(http.MethodPost,relativePath,handlers)
	r.handle(http.MethodDelete,relativePath,handlers)
	r.handle(http.MethodPut,relativePath,handlers)
	r.handle(http.MethodOptions,relativePath,handlers)
	r.handle(http.MethodPatch,relativePath,handlers)
	return	r.handle(http.MethodHead,relativePath,handlers)
}
//自动将struct中暴露的方法注册为ANY
func (r *Route) AUTO(relativePath string, pkg interface{}) Router {
	vpkg := reflect.ValueOf(pkg)
	if !(vpkg.Kind() == reflect.Ptr && reflect.Indirect(vpkg).Kind() == reflect.Struct) {
		panic("must be a struct pointer")
	}

	for i:=0;i<vpkg.NumMethod();i++{
		spath := vpkg.Type().Method(i).Name
		path := path.Join(relativePath,spath)
		r.ANY(path,vpkg.Interface().(tRequestHandlerFunc))
	}

	return r.returnObj()
}

//路由分组
func (r *Route) Group(relativePath string,handlers ...HandlerFunc) Router {
	router := &Route{
		RouteConf: r.RouteConf,
		tree:      r.tree,
		basePath:  r.mergeAbsolutePath(relativePath),
		Handlers:  r.mergeHandlers(handlers),
	}
	return router
}

//静态文件路由
func (r *Route) StaticFile(relative,path string) Router {
	handle := func(c *Context) {
		http.ServeFile(c.Writer,c.Request,path)
	}
	r.GET(relative, handle)
	r.HEAD(relative,handle)
	return r.returnObj()
}

//横向切面，AOP编程
//TODO：更多层次的切面
func (r *Route) Use(handles ...HandlerFunc) Router {
	r.Handlers = append(r.Handlers,handles...)
	return r.returnObj()
}


func (r *Route) Run(rw http.ResponseWriter,req *http.Request) {
	method := req.Method
	path := req.URL.Path
	//这个tsr到底是干啥的
	handlers, ps, _ := r.tree.GetValues(method, path,nil, r.RouteConf.PathUnescape)

	ctx := ctxpool.Get().(*Context)
	ctx.Reset(rw,req,r,handlers,ps)
	ctx.Next()
	ctxpool.Put(ctx)
	//TODO：其他处理
}

func (r *Route)BasePath() string {
	return r.basePath
}

func (r *Route) handle (method,relativePath string,handles []HandlerFunc) Router {
	p := r.mergeAbsolutePath(relativePath)
	chain := r.mergeHandlers(handles)
	r.tree.AddRouter(method,p,chain)
	return r.returnObj()
}

//组合use注册的handle，和传入的handle
func (r *Route)mergeHandlers(chain []HandlerFunc) []HandlerFunc {
	size := len(r.Handlers) + len(chain)
	if  size > int(HandlerLimit) {
		panic("too much handler, limit 63")
	}
	mergedHandlers := make([]HandlerFunc,size)
	copy(mergedHandlers,r.Handlers)
	copy(mergedHandlers,chain)
	return mergedHandlers
}

//合并两个路径
func (r *Route)mergeAbsolutePath(relative string) string {
	return path.Join(r.basePath,relative)
}

func (r *Route) returnObj() Router {
	return r
}



