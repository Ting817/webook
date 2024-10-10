package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Create(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, status uint8) error
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ?", id).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要不 ID 是错的， 要不是创作者不对，特别要注意后者
			return fmt.Errorf("有人在做破坏，误操作非自己的文章。id: %d", id)
		}
		return tx.Model(&PublishedArticle{}).Where("id = ?", id).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			}).Error
	})
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	tx := dao.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	txDAO := NewGORMArticleDAO(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDAO.Create(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle{
		Article: art,
	}
	publishArt.Utime = now
	publishArt.Ctime = now
	err = tx.Clauses(clause.OnConflict{
		// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			//"status":  art.Status,
			"utime": now,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

// SyncClosure 其中 tx => Transaction, trx, txn, 实现 GORM 的事务闭包, GORM 帮助我们管理了事务的生命周期，Begin/Rollback/Commit都不需要我们担心
func (dao *GORMArticleDAO) SyncClosure(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Create(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishArt := PublishedArticle{
			Article: art,
		}
		publishArt.Utime = now
		publishArt.Ctime = now
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   art.Title,
				"content": art.Content,
				"status":  art.Status,
				"utime":   now,
			}),
		}).Create(&publishArt).Error
	})
	return id, err
}

func (dao *GORMArticleDAO) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 只更新标题、内容和状态
func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 依赖 gorm 忽略零值的特性，会用主键id进行更新
	res := dao.db.WithContext(ctx).Model(&art).Where("id=? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		})
	// 检查是否真的更新了
	if res.Error != nil {
		return res.Error
	}
	// res.RowsAffected // 更新行数
	if res.RowsAffected == 0 {
		//dangerousDBOp.Count(1)
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
	}
	return res.Error
}

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement"`
	Title   string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`
	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index"`
	// - 在 author_id 和 ctime 上创建联合索引
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	Status uint8 `gorm:"default=1"`
	Ctime  int64
	Utime  int64
}
