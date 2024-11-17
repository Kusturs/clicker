package main

import (
    "context"
    "fmt"
    "sync"
    "time"
    "sync/atomic"
    "net/http"
    "io"
    "sort"
    "encoding/json"
)

type CounterResponse struct {
    TotalClicks int64 `json:"total_clicks"`
}

func main() {
    fmt.Println("Тест максимальной нагрузки за 1 секунду...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    var (
        wg sync.WaitGroup
        successCount atomic.Int64
        errorCount atomic.Int64
        latencies []time.Duration
        latencyMutex sync.Mutex
    )
    
    // Запускаем горутины
    for i := 0; i < 250; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    start := time.Now()
                    resp, err := http.Get("http://localhost:8080/counter/1")
                    duration := time.Since(start)
                    
                    if err != nil {
                        errorCount.Add(1)
                        continue
                    }
                    
                    // Проверяем статус ответа
                    if resp.StatusCode != http.StatusOK {
                        resp.Body.Close()
                        errorCount.Add(1)
                        continue
                    }
                    
                    // Читаем и проверяем ответ
                    body, err := io.ReadAll(resp.Body)
                    resp.Body.Close()
                    if err != nil {
                        errorCount.Add(1)
                        continue
                    }
                    
                    // Парсим ответ
                    var result CounterResponse
                    if err := json.Unmarshal(body, &result); err != nil {
                        errorCount.Add(1)
                        continue
                    }
                    
                    // Сохраняем латентность
                    latencyMutex.Lock()
                    latencies = append(latencies, duration)
                    latencyMutex.Unlock()
                    
                    successCount.Add(1)
                }
            }
        }()
    }
    
    // Ждем завершения
    <-ctx.Done()
    wg.Wait()
    
    // Анализируем результаты
    fmt.Printf("За 1 секунду:\n")
    fmt.Printf("Успешных запросов: %d\n", successCount.Load())
    fmt.Printf("Ошибок: %d\n", errorCount.Load())
    
    // Считаем статистику по латентности
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })
    
    if len(latencies) > 0 {
        p95 := latencies[len(latencies)*95/100]
        p99 := latencies[len(latencies)*99/100]
        fmt.Printf("Латентность P95: %v\n", p95)
        fmt.Printf("Латентность P99: %v\n", p99)
    }
}
