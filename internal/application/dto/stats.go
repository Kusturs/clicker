package dto

type StatsRequest struct {
    BannerID int64
    TsFrom   int64
    TsTo     int64
}

type StatsResponse struct {
    Stats []StatItem
}

type StatItem struct {
    Timestamp int64 `json:"timestamp"`
    Count     int32 `json:"count"`
}
