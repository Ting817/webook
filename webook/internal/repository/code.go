package repository

import (
	"context"

	"webook/internal/repository/cache"
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(c *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: c,
	}
}

func (repo *CodeRepository) Store(c context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(c, biz, phone, code)
}

func (repo *CodeRepository) Verify(c context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(c, biz, phone, inputCode)
}
