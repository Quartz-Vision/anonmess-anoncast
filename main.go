package main

import (
	"anoncast/server"
	"anoncast/settings"
	"log"
)

func main() {
	if err := settings.Init(); err != nil {
		log.Fatal(err.Error())
		return
	}

	if err := server.Init(); err != nil {
		log.Fatal(err.Error())
		return
	}
}
