package main

import (
    "fmt"
    "net/http"
    "time"
    
    vegeta "github.com/tsenart/vegeta/v12/lib"
)

func runLoadTest(rps int) {
    fmt.Printf("\nТестирование %d RPS:\n", rps)
    fmt.Printf("====================\n")
    
    rate := vegeta.Rate{Freq: rps, Per: time.Second}
    duration := 30 * time.Second
    
    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "GET",
        URL:    "http://localhost:8080/counter/1",
        Header: http.Header{
            "Content-Type": []string{"application/json"},
        },
    })

    attacker := vegeta.NewAttacker(
        vegeta.Timeout(5*time.Second),
        vegeta.Workers(10),
        vegeta.MaxWorkers(20),
    )

    var metrics vegeta.Metrics
    for res := range attacker.Attack(targeter, rate, duration, "Load Test") {
        metrics.Add(res)
    }
    metrics.Close()

    fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
    fmt.Printf("95th percentile: %s\n", metrics.Latencies.P95)
    fmt.Printf("Mean: %s\n", metrics.Latencies.Mean)
    fmt.Printf("Max: %s\n", metrics.Latencies.Max)
    fmt.Printf("Success rate: %.2f%%\n", metrics.Success*100)
    fmt.Printf("Throughput: %.2f RPS\n", metrics.Throughput)
}

func main() {
    // Тестируем разные уровни нагрузки
    loads := []int{500, 1000, 2000, 3000, 4000, 5000}
    
    for _, rps := range loads {
        runLoadTest(rps)
        time.Sleep(5 * time.Second) // Пауза между тестами
    }
}
