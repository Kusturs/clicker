package domain

import "time"

type Click struct {
    ID        int64     `json:"id"`
    BannerID  int64     `json:"banner_id"`
    Timestamp time.Time `json:"timestamp"`
    Count     int       `json:"count"`
}
