package main

import (
	"flag"
	"fmt"
	"log"

	"streamer/api"
	"streamer/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	autoStart := flag.Bool("start", false, "start the stream from saved settings on app launch")
	flag.Parse()

	cfg := config.Load()
	listenAddr := ":" + cfg.ServerPort

	app := fiber.New()
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
