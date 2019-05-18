package bind

import (
	"encoding/json"
	"net/http"
)

const (
	MIME_JSON              = "application/json"
	MIME_HTML              = "text/html"
	MIME_XML               = "application/xml"
	MIME_XML2              = "text/xml"
	MIME_Plain             = "text/plain"
	MIME_POSTForm          = "application/x-www-form-urlencoded"
	MIME_MultipartPOSTForm = "multipart/form-data"
	MIME_PROTOBUF          = "application/x-protobuf"
	MIME_MSGPACK           = "application/x-msgpack"
	MIME_MSGPACK2          = "application/msgpack"
	MIME_YAML              = "application/x-yaml"
)

type Binder interface {
	Name() string
	Parse(req *http.Request,obj interface{}) error
}

var (
	JSON = new(json2)
	PostForm = new(postForm)
	Query = new(query)
	Param = new(param2)
)





type param2 struct {}

func (*param2) Name() string {
	return "Param"
}

func (*param2) Parse(c *ctx.Context,obj interface{}) error {
	bytes, err := json.Marshal(c.params)
	if err != nil{return err}
	return json.Unmarshal(bytes,obj)
}
