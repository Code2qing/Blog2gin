package bootstrap

import (
	"Blog2Gin/conf"
	"Blog2Gin/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
)

func InitDB() {
	databaseConfig := conf.Conf.Database
	gormConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	switch databaseConfig.Type {
	case "mysql":
		{
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				databaseConfig.User, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Name)
			db, err := gorm.Open(mysql.Open(dsn), gormConfig)
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
	default:
		{
			if !(strings.HasSuffix(databaseConfig.DBFile, ".db") && len(databaseConfig.DBFile) > 3) {
				log.Fatalf("db name error.")
			}
			db, err := gorm.Open(sqlite.Open(databaseConfig.DBFile), gormConfig)
			if err != nil {
				log.Fatalf("failed to connect database:%s", err.Error())
			}
			conf.DB = db
			log.Infof("auto migrate model...")
			err = conf.DB.AutoMigrate(&model.Category{}, &model.Tag{}, &model.BlogPost{})
			if err != nil {
				log.Fatalf("failed to auto migrate: %s", err.Error())
			}
		}
	}

}
