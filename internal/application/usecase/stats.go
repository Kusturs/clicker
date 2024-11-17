package usecase

import (
    "context"
    "fmt"
    "log"
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
    
    log.Printf("Getting stats for banner %d from %v to %v", req.BannerID, from, to)
    
    clicks, err := uc.repo.GetStats(ctx, req.BannerID, from, to)
    if err != nil {
        log.Printf("Error getting stats: %v", err)
        return nil, err
    }
    
    var totalClicks int64
    for _, click := range clicks {
        totalClicks += int64(click.Count)
    }
    
    return &dto.StatsResponse{
        TotalClicks: totalClicks,
    }, nil
}
