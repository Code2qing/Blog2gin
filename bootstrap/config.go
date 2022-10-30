package bootstrap

import (
	"Blog2Gin/conf"
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func InitConf() {
	config, err := ioutil.ReadFile(conf.ConfigFile)
	if err != nil {
		log.Fatalf("reading config file error: %s", err.Error())
	}
	err = json.Unmarshal(config, &conf.Conf)
	if err != nil {
		log.Fatalf("load config error: %s", err.Error())
	}
	log.Infof("config: %+v", conf.Conf)
}
