package main

import (
	"anoncast/logging"
	"anoncast/server"
)

func main() {
	if err := server.Start(); err != nil {
		logging.Error.Println(err.Error())
	}
}
