package repository

import (
	"context"
	"webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, r domain.HistoryRecord) error
}
