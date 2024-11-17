package usecase

import (
    "context"
    "testing"
    "time"
    
    "clicker/internal/domain/entity"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockClickRepository struct {
    mock.Mock
}

func (m *MockClickRepository) IncrementClick(ctx context.Context, bannerID int64) (int64, error) {
    args := m.Called(ctx, bannerID)
    return args.Get(0).(int64), args.Error(1)
}

func (m *MockClickRepository) SaveBatch(ctx context.Context, clicks []*entity.Click) error {
    args := m.Called(ctx, clicks)
    return args.Error(0)
}

func (m *MockClickRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    args := m.Called(ctx, bannerID, from, to)
    return args.Get(0).([]*entity.Click), args.Error(1)
}

func TestBatchProcessing(t *testing.T) {
    mockRepo := &MockClickRepository{}
    
    var capturedBatches [][]*entity.Click
    mockRepo.On("SaveBatch", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
        batch := args.Get(1).([]*entity.Click)
        capturedBatches = append(capturedBatches, batch)
    }).Return(nil)
    
    mockRepo.On("IncrementClick", mock.Anything, mock.Anything).Return(int64(1), nil)

    uc := &clickUseCase{
        repo:         mockRepo,
        clickChan:    make(chan *entity.Click, 1000),
        batchSize:    10,
        batchTimeout: 100 * time.Millisecond,
    }
    go uc.processBatch()

    ctx := context.Background()
    
    // Отправляем 25 кликов
    for i := 0; i < 25; i++ {
        _, err := uc.Counter(ctx, 1)
        assert.NoError(t, err)
    }

    // Ждем обработки всех батчей
    time.Sleep(200 * time.Millisecond)

    // Проверяем, что все клики были обработаны батчами
    var totalClicks int
    for _, batch := range capturedBatches {
        totalClicks += len(batch)
    }
    
    assert.Equal(t, 25, totalClicks, "Expected 25 clicks to be processed")
}
