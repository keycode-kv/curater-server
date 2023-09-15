package main

import (
	"curater/app"
	"curater/server"
	"fmt"
)

func main() {
	StartCurater()
}

func StartCurater() {
	err := app.Init()
	if err != nil {
		panic("init failed")
	}

	err = server.Start()
	fmt.Println("error starting server %v", err.Error())

	return
}
