package config

import (
    "fmt"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    PostgresHost     string
    PostgresPort     string
    PostgresUser     string
    PostgresPassword string
    PostgresDB       string

    RestHost string
    RestPort string

    GrpcHost string
    GrpcPort string

    RedisHost     string
    RedisPort     string
    RedisPassword string
    RedisDB       int
}

func New() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, fmt.Errorf("error loading .env file: %w", err)
    }

    return &Config{
        PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
        PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
        PostgresUser:     getEnv("POSTGRES_USER", "clicks_user"),
        PostgresPassword: getEnv("POSTGRES_PASSWORD", "clicks_password"),
        PostgresDB:       getEnv("POSTGRES_DB", "clicks_db"),

        RestHost: getEnv("REST_HOST", "0.0.0.0"),
        RestPort: getEnv("REST_PORT", "8080"),

        GrpcHost: getEnv("GRPC_HOST", "0.0.0.0"),
        GrpcPort: getEnv("GRPC_PORT", "50051"),

        RedisHost:     getEnv("REDIS_HOST", "localhost"),
        RedisPort:     getEnv("REDIS_PORT", "6379"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       0,
    }, nil
}

func (c *Config) GetPostgresDSN() string {
    dsn := os.Getenv("POSTGRES_DSN")
    if dsn != "" {
        return dsn
    }
    
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        c.PostgresUser,
        c.PostgresPassword,
        c.PostgresHost,
        c.PostgresPort,
        c.PostgresDB,
    )
}

func (c *Config) GetRestAddress() string {
    return fmt.Sprintf("%s:%s", c.RestHost, c.RestPort)
}

func (c *Config) GetGrpcAddress() string {
    return fmt.Sprintf("%s:%s", c.GrpcHost, c.GrpcPort)
}

func (c *Config) GetRedisAddress() string {
    return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
