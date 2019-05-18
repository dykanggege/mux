package bind

import (
	"encoding/json"
	"net/http"
)

type json2 struct {
}

//TODO:优化json转化的效率
func (j *json2) Parse(req *http.Request, obj interface{}) error {
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(obj)
}

func (j *json2) Name() string {
	return "JSON"
}
