package main

import "gin"

func main() {
	engine := gin.Default()
	engine.Static()
}