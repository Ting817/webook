package dao

import (
	"gorm.io/gorm"
	"webook/internal/repository/dao/article"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&article.Article{},
		&article.PublishedArticle{},
		&article.PublishedArticleV1{},
		&Interactive{},
		&UserLikeBiz{},
		&Collection{},
		&UserCollectionBiz{},
	) // 若有其他表，则继续往&User{}后添加
}
