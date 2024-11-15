package dto

import "time"

type CounterRequest struct {
    BannerID int64
}

type CounterResponse struct {
    TotalClicks int64
}

type StatsRequest struct {
    BannerID int64
    From     time.Time
    To       time.Time
}

type StatsResponse struct {
    Stats []ClickStats
}

type ClickStats struct {
    Timestamp time.Time
    Count     int32
}
