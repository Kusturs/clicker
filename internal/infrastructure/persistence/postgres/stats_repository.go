package postgres

import (
    "context"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
)

type statsRepository struct {
    db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) repository.StatsRepository {
    return &statsRepository{
        db: db,
    }
}

func (r *statsRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    rows, err := r.db.Query(ctx, `
        SELECT banner_id, date_trunc('hour', timestamp) as hour_timestamp, SUM(count) as total_count
        FROM clicks
        WHERE banner_id = $1 
        AND timestamp >= $2 
        AND timestamp < $3
        GROUP BY banner_id, hour_timestamp
        ORDER BY hour_timestamp
    `, bannerID, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var clicks []*entity.Click
    for rows.Next() {
        click := &entity.Click{}
        if err := rows.Scan(&click.BannerID, &click.Timestamp, &click.Count); err != nil {
            return nil, err
        }
        clicks = append(clicks, click)
    }

    return clicks, rows.Err()
}
