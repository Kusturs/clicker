package app

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "clicker/internal/config"
    "clicker/pkg/clicker"
    "github.com/gorilla/mux"
    "github.com/redis/go-redis/v9"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type App struct {
	cfg    *config.Config
	router *mux.Router
	grpc   *grpc.Server
	redis  *redis.Client
}

func New(cfg *config.Config) *App {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddress(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &App{
		cfg:    cfg,
		router: mux.NewRouter(),
		grpc:   grpc.NewServer(),
		redis:  rdb,
	}
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "REST API is working")
	})

	httpServer := &http.Server{
		Addr:    a.cfg.GetRestAddress(),
		Handler: a.router,
	}

	go func() {
		log.Printf("Starting REST server on %s", a.cfg.GetRestAddress())
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("REST server error: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", a.cfg.GetGrpcAddress())
		if err != nil {
			log.Printf("Failed to listen gRPC: %v", err)
			return
		}
		log.Printf("Starting gRPC server on %s", a.cfg.GetGrpcAddress())
		if err := a.grpc.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	a.grpc.GracefulStop()

	return nil
}

type ClickerService struct {
    clicker.UnimplementedClickerServiceServer
    redis *redis.Client
    clickChan chan *Click
    batchSize int
    batchTimeout time.Duration
}

type Click struct {
    BannerID  int64
    Timestamp time.Time
}

func NewClickerService(redis *redis.Client) *ClickerService {
    s := &ClickerService{
        redis:        redis,
        clickChan:    make(chan *Click, 1000),
        batchSize:    100,
        batchTimeout: 1 * time.Second,
    }
    go s.processBatch()
    return s
}

func (s *ClickerService) processBatch() {
    batch := make([]*Click, 0, s.batchSize)
    ticker := time.NewTicker(s.batchTimeout)
    defer ticker.Stop()

    for {
        select {
        case click := <-s.clickChan:
            batch = append(batch, click)
            if len(batch) >= s.batchSize {
                s.saveBatch(batch)
                batch = make([]*Click, 0, s.batchSize)
            }
        case <-ticker.C:
            if len(batch) > 0 {
                s.saveBatch(batch)
                batch = make([]*Click, 0, s.batchSize)
            }
        }
    }
}

func (s *ClickerService) saveBatch(batch []*Click) {
    ctx := context.Background()
    pipe := s.redis.Pipeline()
    
    for _, click := range batch {
        key := fmt.Sprintf("banner:%d:%d", click.BannerID, click.Timestamp.Unix()/60*60)
        pipe.IncrBy(ctx, key, 1)
    }
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        log.Printf("Error saving batch: %v", err)
    }
}

func (s *ClickerService) IncrementClicks(ctx context.Context, req *clicker.IncrementClicksRequest) (*clicker.IncrementClicksResponse, error) {
    key := fmt.Sprintf("banner:%d:clicks", req.BannerId)
    
    total, err := s.redis.Incr(ctx, key).Result()
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to increment clicks: %v", err)
    }

    return &clicker.IncrementClicksResponse{
        Success: true,
        TotalClicks: total,
    }, nil
}

func (s *ClickerService) GetClickStats(ctx context.Context, req *clicker.GetClickStatsRequest) (*clicker.GetClickStatsResponse, error) {
    fromTime := time.Unix(req.FromTimestamp, 0)
    toTime := time.Unix(req.ToTimestamp, 0)
    
    pipe := s.redis.Pipeline()
    keys := []string{}
    
    // Получаем все ключи для минутных интервалов
    for t := fromTime; t.Before(toTime); t = t.Add(time.Minute) {
        key := fmt.Sprintf("banner:%d:%d", req.BannerId, t.Unix()/60*60)
        keys = append(keys, key)
        pipe.Get(ctx, key)
    }
    
    cmds, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
    }
    
    stats := make([]*clicker.ClickStats, 0, len(cmds))
    var totalClicks int64
    
    for i, cmd := range cmds {
        count := int64(0)
        if cmd.Err() != redis.Nil {
            count, _ = cmd.(*redis.StringCmd).Int64()
        }
        
        if count > 0 {
            stats = append(stats, &clicker.ClickStats{
                Timestamp: fromTime.Add(time.Duration(i) * time.Minute).Unix(),
                Count:    int32(count),
            })
            totalClicks += count
        }
    }
    
    return &clicker.GetClickStatsResponse{
        BannerId:    req.BannerId,
        TotalClicks: totalClicks,
        Stats:       stats,
    }, nil
}
