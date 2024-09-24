package dao

import (
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Article{}) // 若有其他表，则继续往&User{}后添加
}
