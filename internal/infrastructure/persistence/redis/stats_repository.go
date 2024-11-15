package redis

import (
    "context"
    "fmt"
    "time"
    
    "clicker/internal/domain/entity"
    "github.com/redis/go-redis/v9"
)

type statsRepository struct {
    redis *redis.Client
}

func NewStatsRepository(redis *redis.Client) *statsRepository {
    return &statsRepository{
        redis: redis,
    }
}

func (r *statsRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    pipe := r.redis.Pipeline()
    
    var timestamps []time.Time
    for t := from; t.Before(to); t = t.Add(time.Hour) {
        timestamps = append(timestamps, t)
        key := fmt.Sprintf("banner:%d:%d", bannerID, t.Unix()/3600)
        pipe.Get(ctx, key)
    }
    
    cmds, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }
    
    clicks := make([]*entity.Click, 0, len(cmds))
    
    for i, cmd := range cmds {
        count := int64(0)
        if cmd.Err() != redis.Nil {
            count, _ = cmd.(*redis.StringCmd).Int64()
        }
        
        if count > 0 {
            clicks = append(clicks, &entity.Click{
                BannerID:  bannerID,
                Timestamp: timestamps[i],
                Count:    int(count),
            })
        }
    }
    
    return clicks, nil
}
