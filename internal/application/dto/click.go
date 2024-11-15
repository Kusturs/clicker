package dto

import "time"

type CounterRequest struct {
    BannerID int64
}

type CounterResponse struct {
    TotalClicks int64
}

type ClickStats struct {
    Timestamp time.Time
    Count     int32
}
