package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/shivamsahu-tech/event-ingestion/internal/db"
    "github.com/shivamsahu-tech/event-ingestion/internal/queue"
)

func main() {
    ctx := context.Background()

    queue.InitRedis("localhost:6379")
    fmt.Println("Connected to Redis")

    dbURL := "postgres://postgres:postgres@localhost:5432/analytics?sslmode=disable"
    if err := db.ConnectPostgres(dbURL); err != nil {
        panic(err)
    }
    fmt.Println("Connected to Postgres")

    fmt.Println("Worker started... waiting for events")

    for {
        raw, err := queue.PopEvent()
        if err != nil {
            fmt.Println("Redis error:", err)
            continue
        }

        parts := strings.Split(raw, "|")
        if len(parts) != 5 {
            fmt.Println("Invalid event format:", raw)
            continue
        }

        site := parts[0]
        eventType := parts[1]
        path := parts[2]
        userID := parts[3]
        timestamp := parts[4]

        _, err = db.Pool.Exec(ctx,
            `INSERT INTO events (site_id, event_type, path, user_id, timestamp)
             VALUES ($1, $2, $3, $4, $5)`,
            site, eventType, path, userID, timestamp,
        )

        if err != nil {
            fmt.Println("DB insert error:", err)
        } else {
            fmt.Println("Inserted:", eventType, path)
        }
    }
}
