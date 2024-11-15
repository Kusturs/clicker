package usecase

import (
    "context"
    "time"

    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
)

type clickUseCase struct {
    repo      repository.ClickRepository
    clickChan chan *entity.Click
    batchSize int
    batchTimeout time.Duration
}

func NewClickUseCase(repo repository.ClickRepository) repository.ClickUseCase {
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
    uc.clickChan <- &entity.Click{
        BannerID:  bannerID,
        Timestamp: time.Now(),
    }
    return uc.repo.IncrementClick(ctx, bannerID)
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
                uc.repo.SaveBatch(context.Background(), batch)
                batch = make([]*entity.Click, 0, uc.batchSize)
            }
        case <-ticker.C:
            if len(batch) > 0 {
                uc.repo.SaveBatch(context.Background(), batch)
                batch = make([]*entity.Click, 0, uc.batchSize)
            }
        }
    }
}
