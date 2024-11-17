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
    "database/sql"

    "clicker/internal/application/usecase"
    "clicker/internal/config"
    cache "clicker/internal/infrastructure/persistence/redis"
    "clicker/internal/interfaces/grpc/handler"
    "clicker/pkg/counter"
    "clicker/pkg/stats"
    
    "github.com/gorilla/mux"
    "github.com/redis/go-redis/v9"
    "google.golang.org/grpc"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc/credentials/insecure"
    _ "github.com/lib/pq"
)

type App struct {
    cfg    *config.Config
    router *mux.Router
    grpc   *grpc.Server
    redis  *redis.Client
    db     *sql.DB
}

func New(cfg *config.Config) *App {
    db, err := sql.Open("postgres", cfg.GetPostgresDSN())
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }

    rdb := redis.NewClient(&redis.Options{
        Addr:     cfg.GetRedisAddress(),
        Password: cfg.RedisPassword,
        DB:       cfg.RedisDB,
    })

    statsRepo := cache.NewStatsRepository(rdb)
    statsUseCase := usecase.NewStatsUseCase(statsRepo)
    statsHandler := handler.NewStatsHandler(statsUseCase)

    clickRepo := cache.NewClickRepository(rdb)
    clickUseCase := usecase.NewClickUseCase(clickRepo)
    clickHandler := handler.NewClickHandler(clickUseCase)

    grpcServer := grpc.NewServer()
    
    grpcHandler := handler.NewHandler(clickHandler, statsHandler)
    grpcHandler.Register(grpcServer)

    router := mux.NewRouter()
    
    gwmux := runtime.NewServeMux()
    
    opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
    
    if err := counter.RegisterCounterServiceHandlerFromEndpoint(context.Background(), 
        gwmux, cfg.GetGrpcAddress(), opts); err != nil {
        log.Fatalf("Failed to register gateway: %v", err)
    }

    if err := stats.RegisterStatsServiceHandlerFromEndpoint(context.Background(), 
        gwmux, cfg.GetGrpcAddress(), opts); err != nil {
        log.Fatalf("Failed to register gateway: %v", err)
    }

    router.PathPrefix("/").Handler(gwmux)

    return &App{
        cfg:    cfg,
        router: router,
        grpc:   grpcServer,
        redis:  rdb,
        db:     db,
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
    
    if err := a.redis.Close(); err != nil {
        log.Printf("Redis connection close error: %v", err)
    }

    return nil
}
