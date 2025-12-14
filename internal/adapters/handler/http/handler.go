package http

import (
	"github.com/diogomassis/url-shortener/internal/core/ports"
	"github.com/gofiber/fiber/v2"
)

type HTTPHandler struct {
	service ports.URLService
}

func NewHTTPHandler(service ports.URLService) *HTTPHandler {
	return &HTTPHandler{
		service: service,
	}
}

func (h *HTTPHandler) Shorten(c *fiber.Ctx) error {
	var req struct {
		URL string `json:"url"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	url, err := h.service.Shorten(req.URL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(url)
}

func (h *HTTPHandler) Redirect(c *fiber.Ctx) error {
	shortCode := c.Params("shortCode")
	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
	}

	return c.Redirect(originalURL, fiber.StatusMovedPermanently)
}
