package handler

import (
    "context"
    "clicker/internal/application/dto"
    "clicker/internal/application/usecase"
    "clicker/pkg/stats"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type StatsHandler struct {
    stats.UnimplementedStatsServiceServer
    useCase usecase.StatsUseCase
}

func NewStatsHandler(useCase usecase.StatsUseCase) *StatsHandler {
    return &StatsHandler{useCase: useCase}
}

func (h *StatsHandler) Stats(ctx context.Context, req *stats.StatsRequest) (*stats.StatsResponse, error) {
    dtoReq := dto.StatsRequestFromProto(req)
    if dtoReq == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    dtoResp, err := h.useCase.GetStats(ctx, dtoReq)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return dto.ToProtoResponse(dtoResp), nil
}
