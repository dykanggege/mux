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
	return &Mux{}
}

//使用配置文件
func NewConf(path string) *Mux {
	return &Mux{}
}

type Mux struct {
	ctxPool *sync.Pool
}

func (m *Mux) ServeHTTP(rw http.ResponseWriter,req *http.Request) {
	//TODO:从池子中取出 context
	ctx := ctxPool.Get().(*Context)
	ctx.Request = req
	ctx.RW = rw

	//TODO:寻找路由匹配处理
		//TODO:过滤中间件处理
		//TODO:查找路由处理

	ctxPool.Put(ctx)
}
