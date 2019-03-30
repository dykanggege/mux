package bind

import (
	"encoding/json"
	"errors"
	"net/http"
)

type JSON struct {
}

func (j *JSON) Name() string {
	return "json"
}

func (j *JSON) Parse(r *http.Request,obj interface{}) error {
	if r == nil || r.Body == nil{
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(obj)
	return err
}
