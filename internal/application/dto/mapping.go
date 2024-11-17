package dto

import (
    "clicker/pkg/stats"
    "clicker/pkg/counter"
    "clicker/internal/domain/entity"
)

func CounterRequestFromProto(req *counter.CounterRequest) *CounterRequest {
    if req == nil {
        return nil
    }
    return &CounterRequest{
        BannerID: req.BannerId,
    }
}

func ToCounterProtoResponse(resp *CounterResponse) *counter.CounterResponse {
    if resp == nil {
        return nil
    }
    return &counter.CounterResponse{
        TotalClicks: resp.TotalClicks,
    }
}

func StatsRequestFromProto(req *stats.StatsRequest) *StatsRequest {
    if req == nil {
        return nil
    }
    return &StatsRequest{
        BannerID: req.BannerId,
        TsFrom:   req.TsFrom,
        TsTo:     req.TsTo,
    }
}

func ToProtoResponse(resp *StatsResponse) *stats.StatsResponse {
    if resp == nil {
        return nil
    }
    response := &stats.StatsResponse{
        Stats: make([]*stats.StatsResponse_ClickStats, len(resp.Stats)),
    }
    
    for i, stat := range resp.Stats {
        response.Stats[i] = &stats.StatsResponse_ClickStats{
            Timestamp: stat.Timestamp,
            Count:    stat.Count,
        }
    }
    
    return response
}

func FromEntitySlice(clicks []*entity.Click) []StatItem {
    result := make([]StatItem, len(clicks))
    for i, click := range clicks {
        result[i] = StatItem{
            Timestamp: click.Timestamp.Unix(),
            Count:    int32(click.Count),
        }
    }
    return result
}
