package usecase

import (
    "context"
    "fmt"
    "time"
    
    "clicker/internal/application/dto"
    "clicker/internal/domain/repository"
)

type StatsUseCase interface {
    GetStats(ctx context.Context, req *dto.StatsRequest) (*dto.StatsResponse, error)
}

type statsUseCase struct {
    repo repository.StatsRepository
}

func NewStatsUseCase(repo repository.StatsRepository) StatsUseCase {
    return &statsUseCase{
        repo: repo,
    }
}

func (uc *statsUseCase) GetStats(ctx context.Context, req *dto.StatsRequest) (*dto.StatsResponse, error) {
    from := time.Unix(req.TsFrom, 0)
    to := time.Unix(req.TsTo, 0)
    
    if from.After(to) {
        return nil, fmt.Errorf("invalid time range: from is after to")
    }
    
    clicks, err := uc.repo.GetStats(ctx, req.BannerID, from, to)
    if err != nil {
        return nil, err
    }
    
    return &dto.StatsResponse{
        Stats: dto.FromEntitySlice(clicks),
    }, nil
}
