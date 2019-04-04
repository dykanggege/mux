package mux

import "net/http"

type query struct {

}

func (*query) Name() string {
	return "Query"
}

func (*query) Parse(r *http.Request,obj interface{}) error {
	//TODO:query parse

	return nil
}

