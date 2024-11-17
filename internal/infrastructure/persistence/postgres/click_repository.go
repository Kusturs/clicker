package postgres

import (
    "context"
    "database/sql"
    "time"
    
    "clicker/internal/domain/entity"
    "clicker/internal/domain/repository"
)

type clickRepository struct {
    db *sql.DB
}

func NewClickRepository(db *sql.DB) repository.ClickRepository {
    return &clickRepository{
        db: db,
    }
}

func (r *clickRepository) IncrementClick(ctx context.Context, bannerID int64) (int64, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback()

    // Вставляем новый клик
    _, err = tx.ExecContext(ctx, `
        INSERT INTO clicks (banner_id, timestamp, count)
        VALUES ($1, $2, 1)
    `, bannerID, time.Now())
    if err != nil {
        return 0, err
    }

    // Получаем общее количество кликов
    var total int64
    err = tx.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(count), 0)
        FROM clicks
        WHERE banner_id = $1
    `, bannerID).Scan(&total)
    if err != nil {
        return 0, err
    }

    if err = tx.Commit(); err != nil {
        return 0, err
    }

    return total, nil
}

func (r *clickRepository) SaveBatch(ctx context.Context, clicks []*entity.Click) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO clicks (banner_id, timestamp, count)
        VALUES ($1, $2, $3)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, click := range clicks {
        _, err = stmt.ExecContext(ctx, click.BannerID, click.Timestamp, click.Count)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (r *clickRepository) GetStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*entity.Click, error) {
    rows, err := r.db.QueryContext(ctx, `
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
