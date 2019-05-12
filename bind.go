package mux

import (
	"encoding/json"
	"mux/ctx"
)

const (
	mimeJSON              = "application/json"
	mimeHTML              = "text/html"
	mimeXML               = "application/xml"
	mimeXML2              = "text/xml"
	mimePlain             = "text/plain"
	mimePOSTForm          = "application/x-www-form-urlencoded"
	mimeMultipartPOSTForm = "multipart/form-data"
	mimePROTOBUF          = "application/x-protobuf"
	mimeMSGPACK           = "application/x-msgpack"
	mimeMSGPACK2          = "application/msgpack"
	mimeYAML              = "application/x-yaml"
)

var (
	BindJSON = new(json2)
	BindPostForm = new(postForm)
	BindQuery = new(query)
	BindParam = new(param2)
)

type Binding interface {
	Name() string
	Parse(*ctx.Context,interface{}) error
}

type json2 struct {
}

func (j *json2) Parse(*ctx.Context, interface{}) error {
	panic("nothing")

}

func (j *json2) Name() string {
	return "JSON"
}

type query struct {}

func (*query) Name() string {
	return "Query"
}

func (*query) Parse(c *ctx.Context,obj interface{}) error {
	//TODO:bind parse query 性能优化
	if c.querys == nil{
		c.querys = c.Request.URL.Query()
	}
	res := make(map[string]string,len(c.querys))
	for k,v := range c.querys{
		res[k] = v[0]
	}
	bytes, err := json.Marshal(res)
	if err != nil{return err}
	return json.Unmarshal(bytes, obj)
}

type postForm struct {}

func (*postForm) Name() string {
	return "PostForm"
}

func (*postForm) Parse(c *ctx.Context,obj interface{}) (err error) {
	//TODO:bind parse postform 性能优化
	r := c.Request
	err = r.ParseMultipartForm(c.mux.MaxMultipartMemory)
	res := make(map[string]string,len(r.PostForm))
	for key,val := range r.PostForm{
		res[key] = val[0]
	}
	bytes, err := json.Marshal(res)
	if err != nil {return err}
	err = json.Unmarshal(bytes,obj)
	return
}

type param2 struct {}

func (*param2) Name() string {
	return "Param"
}

func (*param2) Parse(c *ctx.Context,obj interface{}) error {
	bytes, err := json.Marshal(c.params)
	if err != nil{return err}
	return json.Unmarshal(bytes,obj)
}
