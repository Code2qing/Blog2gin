package bootstrap

import (
	"Blog2Gin/conf"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() {
	databaseConfig := conf.Conf.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		databaseConfig.User, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database:%s", err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to set db parameters%s", err.Error())
	}
	sqlDB.SetMaxIdleConns(10)
	conf.DB = db
}
