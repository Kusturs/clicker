package repository

import (
    "context"
    "log"
    "sort"
    "time"
    
    "clicker/internal/domain/entity"
)

type compositeStatsRepository struct {
    postgres StatsRepository
    redis    StatsRepository
}

func NewCompositeStatsRepository(postgres, redis StatsRepository) StatsRepository {
    return &compositeStatsRepository{
        postgres: postgres,
        redis:    redis,
    }
}

func (r *compositeStatsRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    boundaryTime := time.Now().Add(-24 * time.Hour)

    var recentClicks, historicalClicks []*entity.Click
    var err error

    if to.After(boundaryTime) {
        recentFrom := from
        if recentFrom.Before(boundaryTime) {
            recentFrom = boundaryTime
        }
        recentClicks, err = r.redis.GetStats(ctx, bannerID, recentFrom, to)
        if err != nil {
            log.Printf("Failed to get recent stats from Redis: %v", err)
        }
    }

    if from.Before(boundaryTime) {
        historicalTo := to
        if historicalTo.After(boundaryTime) {
            historicalTo = boundaryTime
        }
        historicalClicks, err = r.postgres.GetStats(ctx, bannerID, from, historicalTo)
        if err != nil {
            return nil, err
        }
    }

    return mergeClickStats(historicalClicks, recentClicks), nil
}

func mergeClickStats(historical, recent []*entity.Click) []*entity.Click {
    merged := make(map[time.Time]*entity.Click)
    
    for _, click := range historical {
        hourTime := click.Timestamp.Truncate(time.Hour)
        merged[hourTime] = click
    }
    
    for _, click := range recent {
        hourTime := click.Timestamp.Truncate(time.Hour)
        if existing, ok := merged[hourTime]; ok {
            existing.Count += click.Count
        } else {
            merged[hourTime] = click
        }
    }
    
    result := make([]*entity.Click, 0, len(merged))
    for _, click := range merged {
        result = append(result, click)
    }
    
    sort.Slice(result, func(i, j int) bool {
        return result[i].Timestamp.Before(result[j].Timestamp)
    })
    
    return result
}
