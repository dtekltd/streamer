package main

import (
	"fmt"
	"log"

	"streamer/api"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	handler := api.NewHandler()
	handler.RegisterRoutes(app)

	fmt.Println("Web interface starting...")
	fmt.Println("Open your browser and go to: http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
