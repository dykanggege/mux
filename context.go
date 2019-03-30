package mux

import (
	"io"
	"mime/multipart"
	"mux/bind"
	"net/http"
	"os"
)

type bodyJSON struct {

}

//http上下文信息
// 1.参数解析
// 2.格式化返回值功能
// 3.路由传值
type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	//路径参数
	params Params
	//URL参数
	querys map[string][]string
	//上下文参数传递
	keys map[string]interface{}

	mux *Mux
	handlers handleChain
	index    int8
}

func (c *Context) reset()  {
	c.index = -1
	c.params = c.params[0:0]
	c.Key = nil
	c.handlers = nil
	c.querys = nil
}

func (c *Context) File(path string)  {
	http.ServeFile(c.Writer,c.Request,path)
}

func (c *Context) filePath() string {
//	TODO:如果是请求文件，解析请求文件的路径
	return ""
}

func (c *Context) Next()  {
	c.index++
	for c.index<int8(len(c.handlers)){
		c.handlers[c.index](c)
		c.index++
	}
}

//path参数 /user/:id
func (c *Context) Param(key string) string {
	return c.params.ByName(key)
}

func (c *Context) ParamGet(key string) (string,bool) {
	return c.params.Get(key)
}

func (c *Context) Params() Params {
	return c.params
}

//url参数 /user?id=1
func (c *Context) Query(key string) string {
	v, _ := c.QueryGet(key)
	return v
}

func (c *Context) QueryDefault(key,def string) string {
	if v, ok := c.QueryGet(key); ok{
		return v
	}
	return def
}

func (c *Context) QueryGet(key string) (v string,ok bool) {
	if arr, ok := c.QueryArrayGet(key);ok{
		return arr[0],true
	}
	return "",false
}

func (c *Context) QueryArray(key string) []string {
	if arr,ok := c.QueryArrayGet(key);ok{
		return arr
	}
	return []string{}
}

func (c *Context) QueryArrayGet(key string) (arr []string,ok bool) {
	if c.querys == nil{
		c.querys = c.Request.URL.Query()
	}
	arr,ok = c.querys[key]
	return
}

func (c *Context) Querys() map[string][]string {
	if c.querys == nil{
		return c.Request.URL.Query()
	}
	return c.querys
}

//body参数
func (c *Context) PostForm(key string) string {
	v, _ := c.PostFormGet(key)
	return v
}

func (c *Context) PostFormDefault(key,def string) string {
	if v,ok := c.PostFormGet(key);ok{
		return v
	}
	return def
}

func (c *Context) PostFormGet(key string) (string,bool) {
	if arr, ok := c.PostFormArrayGet(key); ok {
		return arr[0],ok
	}
	return "",false
}

func (c *Context) PostFormArray(key string) []string {
	if arr, ok := c.PostFormArrayGet(key);ok{
		return arr
	}
	return []string{}
}

func (c *Context) PostFormArrayGet(key string) (arr []string,ok bool) {
	if c.Request.PostForm == nil{
		_ = c.Request.ParseForm()
	}
	arr,ok = c.Request.PostForm[key]
	return
}

func (c *Context) PostFroms(key string) map[string][]string {
	if c.Request.PostForm == nil{
		_ = c.Request.ParseForm()
	}
	return c.Request.PostForm
}


//查找任何参数
func (c *Context) Value(key string) string {
	v, _ := c.ValueGet(key)
	return v
}

func (c *Context) ValueDefault(key,def string) string {
	if v, ok := c.ValueGet(key);ok{
		return v
	}
	return def
}

func (c *Context) ValueGet(key string) (string,bool) {
	v, ok := c.ParamGet(key)
	if ok{
		return v,ok
	}
	v,ok = c.QueryGet(key)
	if ok {
		return v,ok
	}
	v,ok = c.PostFormGet(key)
	if ok {
		return v,ok
	}
	return "",false
}

func (c *Context) BodyJSON()  {
	
}

func (c *Context) BindWith(obj interface{},b bind.Binding) error {
	return b.Parse(c.Request,obj)
}


//获取文件信息与保存文件
func (c *Context) FileInfo(name string) (*multipart.FileHeader, error) {
	_,header, err := c.Request.FormFile(name)
	return header,err
}

func (c *Context) FileSave(header *multipart.FileHeader,path string) error {
	file, err := header.Open()
	if err != nil{
		return err
	}
	defer file.Close()
	create, err := os.Create(path)
	if err != nil{
		return err
	}
	defer create.Close()
	_, err = io.Copy(create, file)
	return err
}

//传递上下文信息
func (c *Context) Set(key string,val interface{})  {
	if c.keys == nil{
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = val
}

func (c *Context) Get(key string) (interface{},bool) {
	if c.keys == nil{
		return nil,false
	}
	val,ok := c.keys[key]
	return val,ok
}


