package main

import (
	"log"

	"myim/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
