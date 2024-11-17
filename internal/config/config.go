package config

import (
    "fmt"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type PostgresConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Database string
    DSN      string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type ServerConfig struct {
    Host string
    Port string
}

type Config struct {
    Postgres PostgresConfig
    Redis    RedisConfig
    Rest     ServerConfig
    Grpc     ServerConfig
}

func New() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, fmt.Errorf("error loading .env file: %w", err)
    }

    return &Config{
        Postgres: PostgresConfig{
            Host:     getEnv("POSTGRES_HOST", "localhost"),
            Port:     getEnv("POSTGRES_PORT", "5432"),
            User:     getEnv("POSTGRES_USER", "clicks_user"),
            Password: getEnv("POSTGRES_PASSWORD", "clicks_password"),
            Database: getEnv("POSTGRES_DB", "clicks_db"),
            DSN:      getEnv("POSTGRES_DSN", ""),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getEnvAsInt("REDIS_DB", 0),
        },
        Rest: ServerConfig{
            Host: getEnv("REST_HOST", "0.0.0.0"),
            Port: getEnv("REST_PORT", "8080"),
        },
        Grpc: ServerConfig{
            Host: getEnv("GRPC_HOST", "0.0.0.0"),
            Port: getEnv("GRPC_PORT", "50051"),
        },
    }, nil
}

func (c *Config) GetPostgresDSN() string {
    if c.Postgres.DSN != "" {
        return c.Postgres.DSN
    }
    
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        c.Postgres.User,
        c.Postgres.Password,
        c.Postgres.Host,
        c.Postgres.Port,
        c.Postgres.Database,
    )
}

func (c *Config) GetRestAddress() string {
    return fmt.Sprintf("%s:%s", c.Rest.Host, c.Rest.Port)
}

func (c *Config) GetGrpcAddress() string {
    return fmt.Sprintf("%s:%s", c.Grpc.Host, c.Grpc.Port)
}

func (c *Config) GetRedisAddress() string {
    return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value, exists := os.LookupEnv(key); exists {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
