package redis

import (
    "context"
    "fmt"
    "time"
    "strconv"
    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
    "github.com/redis/go-redis/v9"
    "log"
    "strings"
)

type clickRepository struct {
    redis *redis.Client
}

func NewClickRepository(redis *redis.Client) repository.ClickRepository {
    return &clickRepository{
        redis: redis,
    }
}

func (r *clickRepository) SaveBatch(ctx context.Context, clicks []*entity.Click) error {
    pipe := r.redis.Pipeline()
    
    for _, click := range clicks {
        key := fmt.Sprintf("banner:%d:%d", click.BannerID, click.Timestamp.Unix())
        pipe.IncrBy(ctx, key, int64(click.Count))
        pipe.Expire(ctx, key, 24*time.Hour)
    }
    
    _, err := pipe.Exec(ctx)
    return err
}

func (r *clickRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    log.Printf("Redis: Getting stats for banner %d from %v to %v", bannerID, from, to)
    
    // Сначала получим все ключи для данного баннера
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

    // Получаем все значения за раз
    pipe := r.redis.Pipeline()
    for _, key := range keys {
        pipe.Get(ctx, key)
    }
    
    cmds, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        log.Printf("Redis: Error executing pipeline: %v", err)
        return nil, err
    }
    
    clicks := make([]*entity.Click, 0)
    
    for i, cmd := range cmds {
        if cmd.Err() == redis.Nil {
            continue
        }
        
        count, err := cmd.(*redis.StringCmd).Int64()
        if err != nil {
            log.Printf("Redis: Error parsing count from key %s: %v", keys[i], err)
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
    
    return clicks, nil
}
