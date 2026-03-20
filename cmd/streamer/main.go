package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"streamer/api"
	"streamer/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	autoStart := flag.Bool("start", false, "start the stream from saved settings on app launch")
	flag.Parse()

	cfg := config.Load()
	listenAddr := ":" + cfg.ServerPort

	app := fiber.New(fiber.Config{
		AppName: "YouTube Streamer",
	})

	// if cfg.IsDevMode() {
	// 	// Use default logger middleware
	// 	app.Use(logger.New())
	// }

	// Custom middleware for POST
	app.Use(func(c *fiber.Ctx) error {
		if c.Method() == "POST" {
			start := time.Now()

			// Process request
			err := c.Next()

			// Log after request completes
			duration := time.Since(start)
			fmt.Printf("[%s] %s %s - %d - %v\n",
				time.Now().Format("2006-01-02 15:04:05"),
				c.Method(),
				c.Path(),
				c.Response().StatusCode(),
				duration,
			)

			return err
		}

		return c.Next()
	})

	handler := api.NewHandler(cfg)
	handler.RegisterRoutes(app)

	if *autoStart {
		if err := handler.AutoStartFromSavedSettings(); err != nil {
			log.Printf("auto-start failed: %v", err)
		} else {
			log.Println("stream auto-started from saved settings")
		}
	}

	fmt.Println("Web interface starting...")
	fmt.Printf("Open your browser and go to: http://localhost:%s\n", cfg.ServerPort)
	log.Fatal(app.Listen(listenAddr))
}
