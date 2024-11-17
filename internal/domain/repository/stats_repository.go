package repository

import (
	"context"
	"time"
	"clicker/internal/application/dto"
	"clicker/internal/domain/entity"
)

type StatsRepository interface {
	GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error)
}

type StatsUseCase interface {
	GetStats(ctx context.Context, req *dto.StatsRequest) (*dto.StatsResponse, error)
}


