package repository

import (
    "context"
    "time"
	"clicker/internal/domain/entity"
)

type ClickRepository interface {
    SaveBatch(ctx context.Context, clicks []*entity.Click) error
    GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error)
}
