package article

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao/article"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

// CachedArticleAuthorRepository 按照道理，这里也是可以搞缓存的
type CachedArticleAuthorRepository struct {
	dao article.ArticleDAO
}

func NewArticleAuthorRepository(dao article.ArticleDAO) ArticleAuthorRepository {
	return &CachedArticleAuthorRepository{
		dao: dao,
	}
}

func (repo *CachedArticleAuthorRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(art))
}

func (repo *CachedArticleAuthorRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(art))
}

func (repo *CachedArticleAuthorRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
