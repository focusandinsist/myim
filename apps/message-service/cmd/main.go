package main

import (
	"log"

	messageapp "myim/apps/message-service/app"
)

func main() {
	application := messageapp.New()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
