package postgres

import (
    "context"
    "log"
    "time"
    
    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type clickRepository struct {
    db *pgxpool.Pool
}

func NewClickRepository(db *pgxpool.Pool) repository.ClickRepository {
    return &clickRepository{
        db: db,
    }
}

func (r *clickRepository) IncrementClick(ctx context.Context, bannerID int64) (int64, error) {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback(ctx)

    _, err = tx.Exec(ctx, `
        INSERT INTO clicks (banner_id, timestamp, count)
        VALUES ($1, $2, 1)
    `, bannerID, time.Now())
    if err != nil {
        return 0, err
    }

    var total int64
    err = tx.QueryRow(ctx, `
        SELECT COALESCE(SUM(count), 0)
        FROM clicks
        WHERE banner_id = $1
    `, bannerID).Scan(&total)
    if err != nil {
        return 0, err
    }

    if err = tx.Commit(ctx); err != nil {
        return 0, err
    }

    return total, nil
}

func (r *clickRepository) SaveBatch(ctx context.Context, clicks []*entity.Click) error {
    batch := &pgx.Batch{}
    
    for _, click := range clicks {
        batch.Queue(
            "INSERT INTO clicks (banner_id, timestamp, count) VALUES ($1, $2, $3)",
            click.BannerID, click.Timestamp, click.Count,
        )
    }
    
    results := r.db.SendBatch(ctx, batch)
    defer results.Close()
    
    return results.Close()
}

func (r *clickRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    log.Printf("Postgres: Getting stats for banner %d from %v to %v", bannerID, from, to)
    
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
        log.Printf("Postgres: Error querying: %v", err)
        return nil, err
    }
    defer rows.Close()

    var clicks []*entity.Click
    for rows.Next() {
        click := &entity.Click{}
        if err := rows.Scan(&click.BannerID, &click.Timestamp, &click.Count); err != nil {
            log.Printf("Postgres: Error scanning row: %v", err)
            return nil, err
        }
        clicks = append(clicks, click)
    }

    log.Printf("Postgres: Returning %d clicks", len(clicks))
    return clicks, rows.Err()
}
