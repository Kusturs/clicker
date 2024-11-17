package dto

type StatsRequest struct {
    BannerID int64
    TsFrom   int64
    TsTo     int64
}

type StatsResponse struct {
    TotalClicks int64 `json:"total_clicks"`
}
