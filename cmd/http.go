package main

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/diogomassis/url-shortener/internal/adapters/handler/http"
	"github.com/diogomassis/url-shortener/internal/adapters/repository/memory"
	"github.com/diogomassis/url-shortener/internal/core/services"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	prometheus := fiberprometheus.New("url-shortener")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Adapters
	repo := memory.NewMemoryRepository()

	// Services
	service := services.NewURLService(repo)

	// Handlers
	handler := http.NewHTTPHandler(service)

	app.Post("/api/v1", handler.Shorten)
	app.Get("/:shortCode", handler.Redirect)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to URL Shortener!")
	})
	app.Listen(":3000")
}
