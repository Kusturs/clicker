package dto

type CounterRequest struct {
    BannerID int64
}

type CounterResponse struct {
    TotalClicks int64
}
