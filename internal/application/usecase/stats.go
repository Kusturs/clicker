package usecase

import (
    "context"
    "fmt"
    "time"
    
    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
)

type statsUseCase struct {
    repo repository.StatsRepository
}

func NewStatsUseCase(repo repository.StatsRepository) repository.StatsUseCase {
    return &statsUseCase{
        repo: repo,
    }
}

func (uc *statsUseCase) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    if from.After(to) {
        return nil, fmt.Errorf("invalid time range: from is after to")
    }
    return uc.repo.GetStats(ctx, bannerID, from, to)
}
