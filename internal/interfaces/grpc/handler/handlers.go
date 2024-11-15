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
	counterHandler *CounterHandler
}

func NewHandler(counterHandler *CounterHandler) Handler {
	return &GRPCHandler{
		counterHandler: counterHandler,
	}
}

func (h *GRPCHandler) Register(grpcServer *grpc.Server) {
	counter.RegisterCounterServiceServer(grpcServer, h.counterHandler)
	stats.RegisterStatsServiceServer(grpcServer, h.counterHandler)
}
