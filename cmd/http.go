package main

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	prometheus := fiberprometheus.New("url-shortener")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to URL Shortener!")
	})
	app.Listen(":3000")
}
