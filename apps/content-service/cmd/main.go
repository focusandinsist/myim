package main

import (
	"log"
	"myim/apps/content-service/app"
)

func main() {
	a, e := app.New()
	if e != nil {
		log.Fatal(e)
	}
	if e = a.Run(); e != nil {
		log.Fatal(e)
	}
}
