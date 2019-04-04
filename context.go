package mux

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

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


	//解析json数据
	jsonBytes []byte
	jsonResult *gjson.Result

	mux *Mux
	handlers handleChain
	index    int8
}

func (c *Context) reset()  {
	c.index = -1
	c.params = c.params[0:0]
	c.keys = nil
	c.handlers = nil
	c.querys = nil
}

func (c *Context) Next()  {
	c.index++
	for c.index<int8(len(c.handlers)){
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Method() string {
	return c.Request.Method
}

func (c *Context) URI() string {
	querys := c.querys
	path := c.Request.URL.Path + "?"
	for key,val := range querys {
		for i := range val{
			path += key+"="+val[i] + "&"
		}
	}
	return path[:len(path)-1]
}

func (c *Context) URIRaw() string {
	return c.Request.RequestURI
}

func (c *Context)Path() string {
	return c.Request.URL.Path
}

func (c *Context)PathRaw() string {
	p := c.Request.URL.RawPath
	if  p != ""{
		return p
	}
	return c.Request.URL.EscapedPath()
}

func (c *Context) Proto() float64 {
	arr := strings.Split(c.Request.Proto, "/")
	v,_ := strconv.ParseFloat(arr[len(arr)-1],10)
	return v
}

func (c *Context) Headers() http.Header {
	return c.Request.Header
}

func (c *Context) HeaderArr(key string) []string {
	return c.Headers()[key]
}

func (c *Context) HeaderGet(key string) string {
	return c.Headers().Get(key)
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

func (c *Context) BindForm(obj interface{}) error {
	//TODO:bindForm
	return nil
}

func (c *Context) BindPostForm(obj interface{}) error {
	return c.BindWith(obj, PostForm)
}

func (c *Context) BindQuery(obj interface{}) error {
	return c.BindWith(obj, Query)
}

func (c *Context) JSONGet(path ...string) (*gjson.Result,error) {
	err := c.jsonResultAvailable()
	if len(path) == 0{
		return c.jsonResult,err
	}else{
		res := c.jsonResult
		for i := range path{
			temp := res.Get(path[i])
			res = &temp
		}
		return res,err
	}
}

func (c *Context) BindJSON(obj interface{},path ...string) error {
	l := len(path)
	if l == 0{
		if err := c.jsonBytesAvailable(); err != nil{return err}
		return json.Unmarshal(c.jsonBytes,obj)
	}else{
		if err := c.jsonResultAvailable(); err != nil{return err}
		var err error
		for i := range path{
			jn := c.jsonResult.Get(path[i]).String()
			err = json.Unmarshal([]byte(jn), obj)
		}
		if err != nil {return err}
	}
	return nil
}

func (c *Context)jsonResultAvailable() error {
	err := c.jsonBytesAvailable()
	if err != nil{return err}
	if c.jsonResult == nil{
		result := gjson.ParseBytes(c.jsonBytes)
		c.jsonResult = &result
	}
	return nil
}

func (c *Context) jsonBytesAvailable() error {
	if c.jsonBytes == nil{
		jn, err := ioutil.ReadAll(c.Request.Body)
		if err != nil && err != io.EOF{
			return err
		}
		c.jsonBytes = jn
	}
	return nil
}

func (c *Context) BindWith(obj interface{},b Binding) error {
	switch b.Name() {
	case "JSON":
		return c.BindJSON(obj)
	default:
		return b.Parse(c.Request,obj)
	}
}

func (c *Context) Bind(obj interface{}) error {
	return nil
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


//返回信息
func (c *Context) File(path string)  {
	http.ServeFile(c.Writer,c.Request,path)
}

func (c *Context) filePath() string {
	//	TODO:如果是请求文件，解析请求文件的路径
	return ""
}

func (c *Context) WriteJSON(obj interface{}) error {
	bs, err := json.Marshal(obj)
	if err != nil{
		return err
	}
	_, err = c.Writer.Write(bs)
	return err
}

func (c *Context) WriteString()  {

}