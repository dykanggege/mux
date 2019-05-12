package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		fmt.Println(r.URL.EscapedPath())
	})
	http.ListenAndServe(":8080",nil)
}

type bg struct {
	beego.Controller
}

func (b *bg)Get()  {
	b.GetSession()
}