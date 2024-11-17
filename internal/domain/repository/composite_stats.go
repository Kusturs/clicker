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
    log.Printf("Getting stats with boundary time: %v", boundaryTime)

    var recentClicks, historicalClicks []*entity.Click
    var err error

    if to.After(boundaryTime) {
        recentFrom := from
        if recentFrom.Before(boundaryTime) {
            recentFrom = boundaryTime
        }
        log.Printf("Getting recent clicks from Redis for period: %v to %v", recentFrom, to)
        recentClicks, err = r.redis.GetStats(ctx, bannerID, recentFrom, to)
        if err != nil {
            log.Printf("Failed to get recent stats from Redis: %v", err)
        }
        log.Printf("Got %d recent clicks from Redis", len(recentClicks))
    }

    if from.Before(boundaryTime) {
        historicalTo := to
        if historicalTo.After(boundaryTime) {
            historicalTo = boundaryTime
        }
        log.Printf("Getting historical clicks from Postgres for period: %v to %v", from, historicalTo)
        historicalClicks, err = r.postgres.GetStats(ctx, bannerID, from, historicalTo)
        if err != nil {
            log.Printf("Failed to get historical stats from Postgres: %v", err)
            return nil, err
        }
        log.Printf("Got %d historical clicks from Postgres", len(historicalClicks))
    }

    merged := mergeClickStats(historicalClicks, recentClicks)
    log.Printf("Merged stats: %d clicks total", len(merged))
    return merged, nil
}

func mergeClickStats(historical, recent []*entity.Click) []*entity.Click {
    log.Printf("Merging %d historical and %d recent clicks", len(historical), len(recent))
    
    // Группируем по timestamp
    merged := make(map[int64]*entity.Click)
    
    // Добавляем исторические данные
    for _, click := range historical {
        ts := click.Timestamp.Unix()
        if existing, ok := merged[ts]; ok {
            existing.Count += click.Count
        } else {
            merged[ts] = &entity.Click{
                BannerID:  click.BannerID,
                Timestamp: click.Timestamp,
                Count:     click.Count,
            }
        }
    }
    
    // Добавляем недавние данные
    for _, click := range recent {
        ts := click.Timestamp.Unix()
        if existing, ok := merged[ts]; ok {
            existing.Count += click.Count
        } else {
            merged[ts] = &entity.Click{
                BannerID:  click.BannerID,
                Timestamp: click.Timestamp,
                Count:     click.Count,
            }
        }
    }
    
    // Преобразуем map в slice
    result := make([]*entity.Click, 0, len(merged))
    for _, click := range merged {
        result = append(result, click)
    }
    
    // Сортируем по времени
    sort.Slice(result, func(i, j int) bool {
        return result[i].Timestamp.Before(result[j].Timestamp)
    })
    
    log.Printf("Merged %d clicks into %d hourly aggregates", 
        len(historical)+len(recent), len(result))
    
    for _, click := range result {
        log.Printf("Hour: %v, Count: %d", 
            click.Timestamp.Format("2006-01-02 15:04"), click.Count)
    }
    
    return result
}
