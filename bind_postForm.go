package mux

import (
	"encoding/json"
	"net/http"
)

type postForm struct {}

func (*postForm) Name() string {
	return "PostForm"
}

func (*postForm) Parse(r *http.Request,obj interface{}) error {
	//TODO:postform性能提升
	if r.PostForm == nil{
		if err := r.ParseForm(); err != nil {return err}
	}
	res := make(map[string]string)
	for key,val := range r.PostForm{
		res[key] = val[0]
	}
	bytes, err := json.Marshal(res)
	if err != nil {return err}
	return json.Unmarshal(bytes,obj)
}
