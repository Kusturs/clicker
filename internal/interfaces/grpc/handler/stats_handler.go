package handler

import (
    "context"
    "time"

    "clicker/internal/domain/repository"
    "clicker/pkg/stats"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type StatsHandler struct {
    stats.UnimplementedStatsServiceServer
    useCase repository.StatsUseCase
}

func NewStatsHandler(useCase repository.StatsUseCase) *StatsHandler {
    return &StatsHandler{
        useCase: useCase,
    }
}

func (h *StatsHandler) Stats(ctx context.Context, req *stats.StatsRequest) (*stats.StatsResponse, error) {
    if req.TsFrom >= req.TsTo {
        return nil, status.Error(codes.InvalidArgument, "ts_from must be less than ts_to")
    }

    clicks, err := h.useCase.GetStats(ctx, req.BannerId, 
        time.Unix(req.TsFrom, 0), 
        time.Unix(req.TsTo, 0))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
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
