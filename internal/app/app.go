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

    "clicker/internal/application/usecase"
    "clicker/internal/config"
    "clicker/internal/infrastructure/persistence/redis"
    "clicker/internal/infrastructure/persistence/postgres"
    "clicker/internal/domain/repository"
    "clicker/internal/interfaces/grpc/handler"
    "clicker/pkg/counter"
    "clicker/pkg/stats"
    
    "github.com/gorilla/mux"
    goredis "github.com/redis/go-redis/v9"
    "google.golang.org/grpc"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc/credentials/insecure"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ServerManager struct {
    cfg        *config.Config
    httpServer *http.Server
    grpcServer *grpc.Server
    services   *Services
}

type Services struct {
    db    *pgxpool.Pool
    redis *goredis.Client
}

func NewServerManager(cfg *config.Config) (*ServerManager, error) {
    ctx := context.Background()
    
    services, err := initServices(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to init services: %w", err)
    }
    
    handlers := buildHandlers(services)
    servers := buildServers(cfg, handlers)

    return &ServerManager{
        cfg:        cfg,
        httpServer: servers.http,
        grpcServer: servers.grpc,
        services:   services,
    }, nil
}

func initServices(ctx context.Context, cfg *config.Config) (*Services, error) {
    db, err := initDatabase(ctx, cfg)
    if err != nil {
        return nil, err
    }

    redis := initRedis(cfg)

    return &Services{
        db:    db,
        redis: redis,
    }, nil
}

type Repositories struct {
    click repository.ClickRepository
    stats repository.StatsRepository
}

func buildRepositories(services *Services) *Repositories {
    pgClick := postgres.NewClickRepository(services.db)
    pgStats := postgres.NewStatsRepository(services.db)
    redisClick := redis.NewClickRepository(services.redis)
    redisStats := redis.NewStatsRepository(services.redis)

    return &Repositories{
        click: repository.NewCompositeClickRepository(pgClick, redisClick),
        stats: repository.NewCompositeStatsRepository(pgStats, redisStats),
    }
}

type UseCases struct {
    click usecase.ClickUseCase
    stats usecase.StatsUseCase
}

func buildUseCases(repos *Repositories) *UseCases {
    return &UseCases{
        click: usecase.NewClickUseCase(repos.click),
        stats: usecase.NewStatsUseCase(repos.stats),
    }
}

type Servers struct {
    http *http.Server
    grpc *grpc.Server
}

func buildServers(cfg *config.Config, h *handler.Handler) *Servers {
    return &Servers{
        http: buildHTTPServer(cfg, buildGatewayMux(cfg)),
        grpc: buildGRPCServer(h),
    }
}

func buildHandlers(services *Services) *handler.Handler {
    repos := buildRepositories(services)
    useCases := buildUseCases(repos)
    
    clickHandler := handler.NewClickHandler(useCases.click)
    statsHandler := handler.NewStatsHandler(useCases.stats)
    
    return handler.NewHandler(clickHandler, statsHandler)
}

func (m *ServerManager) Run() error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    errChan := make(chan error, 2)
    go m.runHTTPServer(errChan)
    go m.runGRPCServer(errChan)

    return m.handleShutdown(ctx, errChan)
}

func (m *ServerManager) handleShutdown(ctx context.Context, errChan chan error) error {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errChan:
        return fmt.Errorf("server error: %w", err)
    case <-quit:
        return m.gracefulShutdown(ctx)
    }
}

func (m *ServerManager) gracefulShutdown(ctx context.Context) error {
    log.Println("Starting graceful shutdown...")

    shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := m.httpServer.Shutdown(shutdownCtx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
    }

    m.grpcServer.GracefulStop()
    
    if err := m.services.redis.Close(); err != nil {
        log.Printf("Redis connection close error: %v", err)
    }

    m.services.db.Close()

    log.Println("Graceful shutdown completed")
    return nil
}

func initDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
    poolConfig, err := pgxpool.ParseConfig(cfg.GetPostgresDSN())
    if err != nil {
        return nil, fmt.Errorf("failed to parse database config: %w", err)
    }
    
    poolConfig.MaxConns = 50
    poolConfig.MinConns = 10
    poolConfig.MaxConnLifetime = time.Hour
    poolConfig.MaxConnIdleTime = 30 * time.Minute
    
    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    if err = pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return pool, nil
}

func initRedis(cfg *config.Config) *goredis.Client {
    return goredis.NewClient(&goredis.Options{
        Addr:     cfg.GetRedisAddress(),
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })
}

func buildGRPCServer(h *handler.Handler) *grpc.Server {
    server := grpc.NewServer()
    h.Register(server)
    return server
}

func buildHTTPServer(cfg *config.Config, gwmux *runtime.ServeMux) *http.Server {
    router := mux.NewRouter()
    router.HandleFunc("/health", healthCheckHandler)
    router.PathPrefix("/").Handler(gwmux)

    return &http.Server{
        Addr:    cfg.GetRestAddress(),
        Handler: router,
    }
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "REST API is working")
}

func buildGatewayMux(cfg *config.Config) *runtime.ServeMux {
    gwmux := runtime.NewServeMux()
    
    opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
    
    if err := counter.RegisterCounterServiceHandlerFromEndpoint(context.Background(), 
        gwmux, cfg.GetGrpcAddress(), opts); err != nil {
        log.Fatalf("Failed to register counter gateway: %v", err)
    }

    if err := stats.RegisterStatsServiceHandlerFromEndpoint(context.Background(), 
        gwmux, cfg.GetGrpcAddress(), opts); err != nil {
        log.Fatalf("Failed to register stats gateway: %v", err)
    }

    return gwmux
}

func (m *ServerManager) runHTTPServer(errChan chan<- error) {
    log.Printf("Starting HTTP server on %s", m.httpServer.Addr)
    if err := m.httpServer.ListenAndServe(); err != http.ErrServerClosed {
        errChan <- fmt.Errorf("HTTP server error: %w", err)
    }
}

func (m *ServerManager) runGRPCServer(errChan chan<- error) {
    lis, err := net.Listen("tcp", m.cfg.GetGrpcAddress())
    if err != nil {
        errChan <- fmt.Errorf("failed to listen gRPC: %w", err)
        return
    }
    
    log.Printf("Starting gRPC server on %s", m.cfg.GetGrpcAddress())
    if err := m.grpcServer.Serve(lis); err != nil {
        errChan <- fmt.Errorf("gRPC server error: %w", err)
    }
}
