package dto

import "time"

type StatsRequest struct {
    BannerID int64
    From     time.Time
    To       time.Time
}

type StatsResponse struct {
    Stats []StatsItem
}

type StatsItem struct {
    Timestamp time.Time
    Count     int32
}
