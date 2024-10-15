package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"webook/internal/domain"
	"webook/internal/repository/article"
	"webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx *gin.Context, id int64) (domain.Article, error)
}

type articleService struct {
	// 1. 在 service 这一层使用两个 repository
	authorRepo article.ArticleAuthorRepository
	readerRepo article.ArticleReaderRepository

	// 2. 在 repo 里面处理制作库和线上库
	// 1 和 2 是互斥的，不会同时存在
	repo   article.ArticleRepository
	logger logger.LoggerV1
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		repo:   repo,
		logger: l,
	}
}

func NewArticleServiceV1(authorRepo article.ArticleAuthorRepository, readerRepo article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
		logger:     l,
	}
}

func (svc *articleService) Withdraw(ctx context.Context, uid, id int64) error {
	return svc.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	// 设置为未发表
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := svc.update(ctx, art)
		return art.Id, err
	}
	return svc.create(ctx, art)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, art)
}

// PublishV1 基于使用两种 repository 的写法
func (svc *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	// 这一段逻辑其实就是 Save
	if art.Id == 0 {
		id, err = svc.authorRepo.Create(ctx, art)
	} else {
		err = svc.authorRepo.Update(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// 保持制作库和线上库的 ID 是一样的
	art.Id = id
	for i := 0; i < 3; i++ {
		err = svc.readerRepo.Save(ctx, art)
		if err == nil {
			break
		}
		svc.logger.Error("部分失败：保存数据到线上库失败",
			logger.Field{Key: "art_id", Value: id},
			logger.Error(err))
		// 在接入了 metrics 或者 tracing 之后，
		// 这边要进一步记录必要的 DEBUG 信息。
	}
	if err != nil {
		svc.logger.Error("部分失败：保存数据到线上库重试都失败了",
			logger.Field{Key: "art_id", Value: id},
			logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (svc *articleService) create(ctx context.Context, art domain.Article) (int64, error) {
	return svc.repo.Create(ctx, art)
}
func (svc *articleService) update(ctx context.Context, art domain.Article) error {
	return svc.repo.Update(ctx, art)
}

func (svc *articleService) GetPublishedById(ctx *gin.Context, id int64) (domain.Article, error) {
	return svc.repo.GetPublishedById(ctx, id)
}

func (svc *articleService) List(ctx context.Context, uid int64, offset, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, uid, offset, limit)
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}
