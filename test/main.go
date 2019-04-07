package main

import (
	"fmt"
	"net/http"
)

func main() {
	str := "az"
	fmt.Println(str[0])
	fmt.Println(str[1])
	fmt.Println(string(str[0]-32))

	str2 := "AZ"
	fmt.Println(str2[0])
	fmt.Println(str2[1])

	http.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("a","v")
	})
}