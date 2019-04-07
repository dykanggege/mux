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

type M map[string]interface{}

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

func (c *Context) reset(r *http.Request,w http.ResponseWriter)  {
	c.Request = r
	c.Writer = w

	c.index = -1
	c.params = nil
	c.keys = nil
	c.handlers = nil
	c.querys = nil
	c.jsonBytes = nil
	c.jsonResult = nil
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


//TODO:cookie、Sessioner、JWT等机制
func (c *Context) CookieGet(key string) (*http.Cookie, error) {
	return c.Request.Cookie(key)
}

func (c *Context) CookieAdd(coo *http.Cookie)  {
	http.SetCookie(c.Writer,coo)
}

func (c *Context) Session(key string) interface{} {
	return c.mux.Session.Get(key)
}

func (c *Context) SessionSet(key string,val interface{})  {
	c.mux.Session.Set(key,val)
}

func (c *Context) SessionDel(key string)  {
	c.mux.Session.Del(key)
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

//body参数，包含了url参数，但是body参数优先
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

func (c *Context) PostFormArrayGet(key string) ([]string,bool) {
	_ = c.Request.ParseMultipartForm(c.mux.MaxMultipartMemory)
	arr,ok := c.Request.PostForm[key]
	if ok && len(arr) > 0{
		return arr,ok
	}
	return arr,false
}

func (c *Context) PostFroms(key string) map[string][]string {
	_ = c.Request.ParseMultipartForm(c.mux.MaxMultipartMemory)
	return c.Request.PostForm
}


//查找任何参数,param -> query -> postform
func (c *Context) Form(key string) string {
	v, _ := c.FromGet(key)
	return v
}

func (c *Context) FormDefault(key,def string) string {
	if v, ok := c.FromGet(key);ok{
		return v
	}
	return def
}

func (c *Context) FromGet(key string) (string,bool) {
	v,ok := c.PostFormGet(key)
	if ok {
		return v,ok
	}
	v,ok = c.QueryGet(key)
	if ok {
		return v,ok
	}
	v, ok = c.ParamGet(key)
	if ok{
		return v,ok
	}
	return "",false
}

func (c *Context) BindPostForm(obj interface{}) error {
	return c.BindWith(obj, BindPostForm)
}

func (c *Context) BindQuery(obj interface{}) error {
	return c.BindWith(obj, BindQuery)
}

func (c *Context) BindParam(obj interface{}) error {
	return c.BindWith(obj,BindParam)
}

//postform > query > param
func (c *Context) BindForm(obj interface{}) error {
	err := c.BindParam(obj)
	err = c.BindQuery(obj)
	err = c.BindPostForm(obj)
	return err
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
		return b.Parse(c,obj)
	}
}

func (c *Context) Bind(obj interface{}) (err error) {
	t := strings.Split(c.Request.Header.Get("Content-Type"),";")[0]
	switch t {
	case mimeJSON:
		err = c.BindJSON(obj)
	case mimePOSTForm,mimeMultipartPOSTForm:
		err = c.BindPostForm(obj)
	default:
		err = c.BindForm(obj)
	}
	return
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
	return c.Param("filepath")
}

func (c *Context) WriteJSON(code int,obj interface{}) error {
	c.Writer.WriteHeader(code)
	bs, err := json.Marshal(obj)
	if err != nil{
		return err
	}
	_, err = c.Writer.Write(bs)
	return err
}

func (c *Context) WriteString(code int,str string)  {
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(str))
}

