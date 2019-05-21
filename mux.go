package mux

import (
	"mux/route"
	"mux/session"
	"net/http"
)



func Default() *Mux {
	m := &Mux{}
	m.Route = route.DefaultRoute
	return m
}


type Mux struct {
	route.Route
	sessionManager session.Manager
}

//TODO:限制连接的最大数量
//TODO:支持fasthttp
func NewMux(config *route.Config) *Mux {
	m := &Mux{}
	m.RouteConf = config
	return m
}

func (m *Mux) ServeHTTP(rw http.ResponseWriter,req *http.Request) {
	m.Route.Run(rw,req)
}

func (m *Mux) Run(port ...string) error {
	l := len(port)
	if l == 0{
		port = append(port,":80")
	}
	return http.ListenAndServe(port[0],m)
}

func (m *Mux) RunTSL(certFile, keyFile string,port ...string) error {
	l := len(port)
	if l == 0{
		port = append(port,":443")
	}
	return http.ListenAndServeTLS(port[0],certFile, keyFile ,m)
}
