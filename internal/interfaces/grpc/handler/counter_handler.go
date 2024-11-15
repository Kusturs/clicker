package handler

import (
    "context"
    "time"

    "clicker/internal/domain/repository"
    "clicker/pkg/counter"
    "clicker/pkg/stats"
)

type CounterHandler struct {
    counter.UnimplementedCounterServiceServer
    stats.UnimplementedStatsServiceServer
    useCase repository.ClickUseCase
}

func NewCounterHandler(useCase repository.ClickUseCase) *CounterHandler {
    return &CounterHandler{useCase: useCase}
}

func (h *CounterHandler) Counter(ctx context.Context, req *counter.CounterRequest) (*counter.CounterResponse, error) {
    total, err := h.useCase.Counter(ctx, req.BannerId)
    if err != nil {
        return nil, err
    }
    return &counter.CounterResponse{TotalClicks: total}, nil
}

func (h *CounterHandler) Stats(ctx context.Context, req *stats.StatsRequest) (*stats.StatsResponse, error) {
    clicks, err := h.useCase.Stats(ctx, req.BannerId, 
        time.Unix(req.TsFrom, 0), 
        time.Unix(req.TsTo, 0))
    if err != nil {
        return nil, err
    }

    response := &stats.StatsResponse{
        Stats: make([]*stats.StatsResponse_ClickStats, len(clicks)),
    }
    for i, click := range clicks {
        response.Stats[i] = &stats.StatsResponse_ClickStats{
            Timestamp: click.Timestamp.Unix(),
            Count:    int32(click.Count),
        }
    }
    return response, nil
}
