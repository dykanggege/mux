package mux

import (
	"net/http"
)

type json2 struct {
}

func (j *json2) Name() string {
	return "JSON"
}

func (j *json2) Parse(r *http.Request,obj interface{}) error {
	panic("nothing")
}
