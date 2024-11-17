package handler

import (
	"clicker/pkg/counter"
	"clicker/pkg/stats"
	"google.golang.org/grpc"
)

type GRPCServer interface {
	Register(*grpc.Server)
}

type ClickService interface {
	counter.CounterServiceServer
}

type StatsService interface {
	stats.StatsServiceServer
}

type Handler struct {
	clickService ClickService
	statsService StatsService
}

func NewHandler(clickService ClickService, statsService StatsService) *Handler {
	return &Handler{
		clickService: clickService,
		statsService: statsService,
	}
}

func (h *Handler) Register(server *grpc.Server) {
	counter.RegisterCounterServiceServer(server, h.clickService)
	stats.RegisterStatsServiceServer(server, h.statsService)
}
