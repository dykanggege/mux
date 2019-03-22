package main

import (
	"fmt"
	"reflect"
)

type face func(s string)

type Iface = func(s string)

type t struct {
}

func (t *t) FuncNameee(s string)  {
	fmt.Println("666")
}

func main() {
	judge(1)
}

func judge(i interface{})  {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr && reflect.Indirect(val).Kind() == reflect.Struct {
		fmt.Println("yes")
	}else{
		fmt.Println("no")
	}
}