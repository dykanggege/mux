package bind

import (
	"encoding/json"
	"net/http"
)

type query struct {}

func (*query) Name() string {
	return "Query"
}

func (*query) Parse(req *http.Request,obj interface{}) error{
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