package bind

import (
	"encoding/json"
	"net/http"
)

type postForm struct {}

func (*postForm) Name() string {
	return "PostForm"
}

func (p *postForm) Parse(r *http.Request,obj interface{}) error {
	//TODO:bind parse postform 性能优化
	err := r.ParseMultipartForm(4 << 20)
	res := make(map[string]string,len(r.PostForm))
	for key,val := range r.PostForm{
		res[key] = val[0]
	}
	bytes, err := json.Marshal(res)
	if err != nil {return err}
	err = json.Unmarshal(bytes,obj)
	return err
}