package main

import (
	"log"

	messageapp "myim/apps/message-service/app"
)

func main() {
	application, err := messageapp.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
