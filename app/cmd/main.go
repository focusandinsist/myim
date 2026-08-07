package main

import (
	"log"

	"myim/app/config"
	"myim/app/dao"
	"myim/app/http"
	"myim/app/service"
)

func main() {
	conf := config.New()
	dao, err := dao.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	defer dao.Close()
	serv := service.New(conf, dao)
	router := http.Init(serv)
	if err := http.Run(conf.Addr, router); err != nil {
		log.Fatal(err)
	}
}
