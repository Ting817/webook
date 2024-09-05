package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/pkg/cfg"

	"webook/internal/repository/dao"
)

func InitDB(c cfg.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.DB.DSN), &gorm.Config{})
	if err != nil {
		panic(err) // panic 即goroutine直接结束 只在初始化时考虑panic
	}
	if err = dao.InitTable(db); err != nil {
		panic(err)
	}

	return db
}
