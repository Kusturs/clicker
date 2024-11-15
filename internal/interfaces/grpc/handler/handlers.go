package handler

import (
	"clicker/pkg/counter"
	"clicker/pkg/stats"
	"google.golang.org/grpc"
)

type Handler interface {
	Register(*grpc.Server)
}

type GRPCHandler struct {
	clickHandler *ClickHandler
	statsHandler *StatsHandler
}

func NewHandler(clickHandler *ClickHandler, statsHandler *StatsHandler) Handler {
	return &GRPCHandler{
		clickHandler: clickHandler,
		statsHandler: statsHandler,
	}
}

func (h *GRPCHandler) Register(grpcServer *grpc.Server) {
	counter.RegisterCounterServiceServer(grpcServer, h.clickHandler)
	stats.RegisterStatsServiceServer(grpcServer, h.statsHandler)
}
