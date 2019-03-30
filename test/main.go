package main

import (
	"fmt"
	"github.com/tidwall/gjson"
)

type N struct {
	Name string `json:"name"`
}

func main() {
	parse := gjson.Parse(`{"name":{"first":"Janet","last":"Prichard"},"age":47}`)
	n := parse.Get("name.first")
	fmt.Println(n)

}