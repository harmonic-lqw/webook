package repository

import (
	"context"
	"webook/internal/domain"
)

type HistoryRecordRepository interface {
	BatchAddRecord(ctx context.Context, record []domain.HistoryRecord) error
}
