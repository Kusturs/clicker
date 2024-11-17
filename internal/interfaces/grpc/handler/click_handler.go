package handler

import (
    "context"
    "clicker/internal/application/usecase"
    "clicker/pkg/counter"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type ClickHandler struct {
    counter.UnimplementedCounterServiceServer
    useCase usecase.ClickUseCase
}

func NewClickHandler(useCase usecase.ClickUseCase) *ClickHandler {
    return &ClickHandler{useCase: useCase}
}

func (h *ClickHandler) Counter(ctx context.Context, req *counter.CounterRequest) (*counter.CounterResponse, error) {
    total, err := h.useCase.Counter(ctx, req.BannerId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return &counter.CounterResponse{
        TotalClicks: total,
    }, nil
}
