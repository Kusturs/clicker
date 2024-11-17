package repository

import (
    "context"
    "log"
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
    boundaryTime := time.Now().Add(-24 * time.Hour)

    if to.After(boundaryTime) {
        clicks, err := r.redis.GetStats(ctx, bannerID, from, to)
        if err == nil && len(clicks) > 0 {
            return clicks, nil
        }
    }

    return r.postgres.GetStats(ctx, bannerID, from, to)
}
