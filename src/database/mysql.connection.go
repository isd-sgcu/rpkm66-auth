package database

import (
	"fmt"
	"github.com/isd-sgcu/rpkm66-auth/src/app/model/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
)

func InitDatabase(conf *config.Database) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", conf.User, conf.Password, conf.Host, strconv.Itoa(conf.Port), conf.Name)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(auth.Auth{})
	if err != nil {
		return nil, err
	}

	return
}
