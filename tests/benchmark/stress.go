package main

import (
    "fmt"
    "net/http"
    "time"
    
    vegeta "github.com/tsenart/vegeta/v12/lib"
)

func runStressTest(rps int) (success float64, latency time.Duration) {
    rate := vegeta.Rate{Freq: rps, Per: time.Second}
    duration := 10 * time.Second
    
    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "GET",
        URL:    "http://localhost:8080/counter/1",
        Header: http.Header{
            "Content-Type": []string{"application/json"},
        },
    })

    attacker := vegeta.NewAttacker(
        vegeta.Timeout(5*time.Second),
        vegeta.Workers(uint64(rps/50)),
        vegeta.MaxWorkers(uint64(rps/25)),
    )

    var metrics vegeta.Metrics
    for res := range attacker.Attack(targeter, rate, duration, "Stress Test") {
        metrics.Add(res)
    }
    metrics.Close()

    return metrics.Success * 100, metrics.Latencies.P95
}

func main() {
    fmt.Println("Начинаем стресс-тестирование...")
    
    rps := 1000
    step := 1000
    maxRps := 0
    
    for {
        fmt.Printf("\nТестирование %d RPS...\n", rps)
        success, latency := runStressTest(rps)
        
        fmt.Printf("Успешность: %.2f%%\n", success)
        fmt.Printf("Латентность (P95): %s\n", latency)
        
        if success < 99 || latency > 100*time.Millisecond {
            maxRps = rps - step
            break
        }
        
        rps += step
        
        time.Sleep(5 * time.Second)
    }
    
    fmt.Printf("\n=== Результаты тестирования ===\n")
    fmt.Printf("Максимальная стабильная нагрузка: %d RPS\n", maxRps)
    
    optimalRps := int(float64(maxRps) * 0.9)
    fmt.Printf("\nПроверка оптимальной нагрузки (%d RPS)...\n", optimalRps)
    success, latency := runStressTest(optimalRps)
    
    fmt.Printf("\n=== Рекомендуемые параметры ===\n")
    fmt.Printf("Оптимальная нагрузка: %d RPS\n", optimalRps)
    fmt.Printf("Успешность при оптимальной нагрузке: %.2f%%\n", success)
    fmt.Printf("Латентность при оптимальной нагрузке (P95): %s\n", latency)
}
