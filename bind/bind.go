package bind

import (
	"net/http"
)

type Binding interface {
	Name() string
	Parse(*http.Request,interface{}) error
}


