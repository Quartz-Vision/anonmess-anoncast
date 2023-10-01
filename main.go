package main

import (
	"anoncast/server"

	"github.com/Quartz-Vision/golog"
)

func main() {
	if err := server.Start(); err != nil {
		golog.Error.Println(err.Error())
	}
}
