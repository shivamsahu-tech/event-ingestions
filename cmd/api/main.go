package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/shivamsahu-tech/event-ingestion/internal/db"
	"github.com/shivamsahu-tech/event-ingestion/internal/models"
	"github.com/shivamsahu-tech/event-ingestion/internal/queue"
)

func main() {

	queue.InitRedis("localhost:6379")

	db.ConnectPostgres("postgres://postgres:postgres@localhost:5432/analytics?sslmode=disable")

	fmt.Println("Ingestion + Reporting API running on :8080")

	app := fiber.New()

	app.Post("/event", func(c *fiber.Ctx) error {
		var ev models.Event

		if err := c.BodyParser(&ev); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}

		if ev.SiteID == "" || ev.EventType == "" {
			return c.Status(400).JSON(fiber.Map{"error": "site_id and event_type required"})
		}

		if ev.Timestamp == "" {
			ev.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		eventStr := fmt.Sprintf(
			"%s|%s|%s|%s|%s",
			ev.SiteID,
			ev.EventType,
			ev.Path,
			ev.UserID,
			ev.Timestamp,
		)

		fmt.Println("Received Event:", eventStr)

		if err := queue.PushEvent(eventStr); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to push to queue"})
		}

		fmt.Println("Event pushed to Redis queue")

		return c.SendStatus(204)
	})


	app.Get("/stats", func(c *fiber.Ctx) error {
		siteID := c.Query("site_id")
		date := c.Query("date") 

		if siteID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "site_id is required"})
		}

		if date == "" {
			return c.Status(400).JSON(fiber.Map{"error": "date is required (YYYY-MM-DD)"})
		}

		ctx := context.Background()

		var totalViews int
		err := db.Pool.QueryRow(ctx,
			`SELECT COUNT(*) 
			 FROM events 
			 WHERE site_id = $1 AND DATE(timestamp) = $2`,
			siteID, date).Scan(&totalViews)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error (count)"})
		}

		var uniqueUsers int
		err = db.Pool.QueryRow(ctx,
			`SELECT COUNT(DISTINCT user_id)
			 FROM events
			 WHERE site_id = $1 AND DATE(timestamp) = $2`,
			siteID, date).Scan(&uniqueUsers)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error (unique users)"})
		}

		rows, err := db.Pool.Query(ctx,
			`SELECT path, COUNT(*) AS views
			 FROM events
			 WHERE site_id = $1 AND DATE(timestamp) = $2
			 GROUP BY path
			 ORDER BY views DESC
			 LIMIT 10`,
			siteID, date)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error (top paths)"})
		}
		defer rows.Close()

		type TopPath struct {
			Path  string `json:"path"`
			Views int    `json:"views"`
		}

		var topPaths []TopPath
		for rows.Next() {
			var p TopPath
			rows.Scan(&p.Path, &p.Views)
			topPaths = append(topPaths, p)
		}

		return c.JSON(fiber.Map{
			"site_id":      siteID,
			"date":         date,
			"total_views":  totalViews,
			"unique_users": uniqueUsers,
			"top_paths":    topPaths,
		})
	})

	app.Listen(":8080")
}
