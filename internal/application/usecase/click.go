package usecase

import (
    "context"
    "time"
    "log"

    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
)

type ClickUseCase interface {
    Counter(ctx context.Context, bannerID int64) (int64, error)
    Stats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error)
}

type clickUseCase struct {
    repo         repository.ClickRepository
    clickChan    chan *entity.Click
    batchSize    int
    batchTimeout time.Duration
}

func NewClickUseCase(repo repository.ClickRepository) ClickUseCase {
    uc := &clickUseCase{
        repo:         repo,
        clickChan:    make(chan *entity.Click, 1000),
        batchSize:    100,
        batchTimeout: time.Second,
    }
    go uc.processBatch()
    return uc
}

func (uc *clickUseCase) Counter(ctx context.Context, bannerID int64) (int64, error) {
    select {
    case uc.clickChan <- &entity.Click{
        BannerID:  bannerID,
        Timestamp: time.Now(),
        Count:     1,
    }:
        now := time.Now()
        from := now.Add(-24 * time.Hour)
        clicks, err := uc.repo.GetStats(ctx, bannerID, from, now)
        if err != nil {
            log.Printf("Failed to get stats: %v", err)
            return 1, nil
        }
        
        var total int64 = 1
        for _, click := range clicks {
            total += int64(click.Count)
        }
        return total, nil
        
    case <-ctx.Done():
        return 0, ctx.Err()
    }
}

func (uc *clickUseCase) Stats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    return uc.repo.GetStats(ctx, bannerID, from, to)
}

func (uc *clickUseCase) processBatch() {
    batch := make([]*entity.Click, 0, uc.batchSize)
    ticker := time.NewTicker(uc.batchTimeout)
    defer ticker.Stop()

    for {
        select {
        case click := <-uc.clickChan:
            batch = append(batch, click)
            if len(batch) >= uc.batchSize {
                if err := uc.saveBatch(batch); err != nil {
                    log.Printf("Failed to save batch: %v", err)
                }
                batch = make([]*entity.Click, 0, uc.batchSize)
            }
        case <-ticker.C:
            if len(batch) > 0 {
                if err := uc.saveBatch(batch); err != nil {
                    log.Printf("Failed to save batch: %v", err)
                }
                batch = make([]*entity.Click, 0, uc.batchSize)
            }
        }
    }
}

func (uc *clickUseCase) saveBatch(batch []*entity.Click) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return uc.repo.SaveBatch(ctx, batch)
}
