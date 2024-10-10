package article

import (
	"context"
	"gorm.io/gorm"
	"webook/internal/domain"
	dao "webook/internal/repository/dao/article"
)

// repository 还是要用来操作缓存和DAO，事物概念应该在 DAO 这一层

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, status domain.ArticleStatus) error
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func NewArticleRepositoryV1(authorDAO dao.ArticleAuthorDAO, readerDAO dao.ArticleReaderDAO) ArticleRepository {
	return &CachedArticleRepository{
		authorDAO: authorDAO,
		readerDAO: readerDAO,
	}
}

type CachedArticleRepository struct {
	// 操作单一的库
	dao dao.ArticleDAO

	// SyncV1 用
	authorDAO dao.ArticleAuthorDAO
	readerDAO dao.ArticleReaderDAO

	// SyncV2 用, 同库不同表
	db *gorm.DB
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = c.authorDAO.Create(ctx, artn)
		if err != nil {
			return 0, err
		}
	} else {
		err = c.authorDAO.UpdateById(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, status.ToUint8())
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Create(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	}
}
