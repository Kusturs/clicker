package handler

import (
    "context"
    "time"

    "clicker/internal/domain/repository"
    "clicker/pkg/counter"
    "clicker/pkg/stats"
)

type ClickHandler struct {
    counter.UnimplementedCounterServiceServer
    stats.UnimplementedStatsServiceServer
    useCase repository.ClickUseCase
}

func NewClickHandler(useCase repository.ClickUseCase) *ClickHandler {
    return &ClickHandler{useCase: useCase}
}

func (h *ClickHandler) Counter(ctx context.Context, req *counter.CounterRequest) (*counter.CounterResponse, error) {
    total, err := h.useCase.Counter(ctx, req.BannerId)
    if err != nil {
        return nil, err
    }
    return &counter.CounterResponse{TotalClicks: total}, nil
}

func (h *ClickHandler) Stats(ctx context.Context, req *stats.StatsRequest) (*stats.StatsResponse, error) {
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
