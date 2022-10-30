package conf

import "gorm.io/gorm"

var (
	ConfigFile string = "conf/app.json"
	Conf       Config
	DB         *gorm.DB
)

const (
	IndexPageSize = 5
)
