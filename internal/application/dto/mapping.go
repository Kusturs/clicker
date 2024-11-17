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
    return &stats.StatsResponse{
        TotalClicks: resp.TotalClicks,
    }
}

func TotalClicksFromEntity(clicks []*entity.Click) int64 {
    var total int64
    for _, click := range clicks {
        total += int64(click.Count)
    }
    return total
}
