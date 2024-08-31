package repository

import (
	"context"

	"webook/internal/repository/cache"
)

type CodeRepository interface {
	Store(c context.Context, biz string, phone string, code string) error
	Verify(c context.Context, biz, phone, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

func (repo *CachedCodeRepository) Store(c context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(c, biz, phone, code)
}

func (repo *CachedCodeRepository) Verify(c context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(c, biz, phone, inputCode)
}
