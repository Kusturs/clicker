package repository

import (
    "context"
    "log"
    "sync"
    "time"
    
    "clicker/internal/domain/entity"
)

type compositeClickRepository struct {
    postgres ClickRepository
    redis    ClickRepository
}

func NewCompositeClickRepository(postgres, redis ClickRepository) ClickRepository {
    return &compositeClickRepository{
        postgres: postgres,
        redis:    redis,
    }
}

func (r *compositeClickRepository) SaveBatch(ctx context.Context, clicks []*entity.Click) error {
    if err := r.postgres.SaveBatch(ctx, clicks); err != nil {
        return err
    }

    if err := r.redis.SaveBatch(ctx, clicks); err != nil {
        log.Printf("Failed to update Redis cache: %v", err)
    }

    return nil
}

func (r *compositeClickRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    var (
        redisClicks, pgClicks []*entity.Click
        redisErr, pgErr error
        wg sync.WaitGroup
    )

    if to.After(time.Now().Add(-24 * time.Hour)) {
        wg.Add(1)
        go func() {
            defer wg.Done()
            redisClicks, redisErr = r.redis.GetStats(ctx, bannerID, from, to)
        }()
    }

    wg.Add(1)
    go func() {
        defer wg.Done()
        pgClicks, pgErr = r.postgres.GetStats(ctx, bannerID, from, to)
    }()

    wg.Wait()
    
    if redisErr == nil && len(redisClicks) > 0 {
        return redisClicks, nil
    }
    return pgClicks, pgErr
}
