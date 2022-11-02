package main

import (
	"Blog2Gin/bootstrap"
	"Blog2Gin/conf"
	"Blog2Gin/server"
	"Blog2Gin/templates"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Init() {
	bootstrap.InitConf()
	bootstrap.InitDB()
}

func main() {
	Init()
	r := gin.Default()
	pprof.Register(r)
	server.InitRouter(r)
	templates.InitTemplate(r)
	base := fmt.Sprintf("%s:%d", conf.Conf.Address, conf.Conf.Port)
	err := r.Run(base)
	if err != nil {
		log.Errorf("failed to start: %s", err.Error())
	}
}
