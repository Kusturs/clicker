package redis

import (
    "context"
    "fmt"
    "log"
    "strconv"
    "strings"
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
    log.Printf("Redis: Getting stats for banner %d from %v to %v", bannerID, from, to)
    
    pattern := fmt.Sprintf("banner:%d:*", bannerID)
    keys, err := r.redis.Keys(ctx, pattern).Result()
    if err != nil {
        log.Printf("Redis: Error getting keys: %v", err)
        return nil, err
    }
    
    log.Printf("Redis: Found %d keys for banner %d", len(keys), bannerID)
    
    if len(keys) == 0 {
        return nil, nil
    }

    pipe := r.redis.Pipeline()
    for _, key := range keys {
        pipe.Get(ctx, key)
    }
    
    cmds, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }

    clicks := make([]*entity.Click, 0)
    for i, cmd := range cmds {
        if cmd.Err() == redis.Nil {
            continue
        }
        
        count, err := cmd.(*redis.StringCmd).Int64()
        if err != nil {
            continue
        }

        parts := strings.Split(keys[i], ":")
        if len(parts) != 3 {
            continue
        }
        
        ts, err := strconv.ParseInt(parts[2], 10, 64)
        if err != nil {
            continue
        }

        timestamp := time.Unix(ts, 0)
        if timestamp.Before(from) || timestamp.After(to) {
            continue
        }

        clicks = append(clicks, &entity.Click{
            BannerID:  bannerID,
            Timestamp: timestamp,
            Count:     int(count),
        })
    }

    log.Printf("Redis: Returning %d clicks", len(clicks))
    return clicks, nil
}
