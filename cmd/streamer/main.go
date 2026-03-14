package main

import (
	"fmt"
	"log"

	"streamer/api"
	"streamer/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()
	listenAddr := ":" + cfg.ServerPort

	app := fiber.New()
	handler := api.NewHandler(cfg)
	handler.RegisterRoutes(app)

	fmt.Println("Web interface starting...")
	fmt.Printf("Open your browser and go to: http://localhost:%s\n", cfg.ServerPort)
	log.Fatal(app.Listen(listenAddr))
}
