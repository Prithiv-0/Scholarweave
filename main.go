package main

import (
	"Scholarweave/api/handlers"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	// Create new Fiber instance with error handling
	app := fiber.New(fiber.Config{
		AppName: "ScholarWeave API v1",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":  err.Error(),
				"status": code,
				"time":   time.Now().Format(time.RFC3339),
				"path":   c.Path(),
				"method": c.Method(),
			})
		},
	})

	//middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
	}))

	//routes
	setupRoutes(app)

	log.Printf("Starting ScholarWeave API on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func setupRoutes(app *fiber.App) {
	// API v1 group
	v1 := app.Group("/api/v1")

	// Initialize handlers
	openAlexHandler := handlers.NewOpenAlexHandler()

	// Root route
	app.Get("/", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Welcome to ScholarWeave API",
			"endpoints": []string{
				"/api/v1/health",
				"/api/v1/papers/search?q=your_query",
				"/api/v1/papers/:id",
			},
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Health check with detailed status
	v1.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
			"services": fiber.Map{
				"api":      "healthy",
				"openalex": "connected",
			},
		})
	})

	// Search papers endpoint with error handling
	v1.Get("/papers/search", func(c fiber.Ctx) error {
		if c.Query("q") == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Search query is required")
		}
		return openAlexHandler.SearchPapers(c)
	})

	// Get paper by ID with validation
	v1.Get("/papers/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Paper ID is required")
		}
		return openAlexHandler.GetPaperByID(c)
	})

	// Handle 404 Not Found
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Endpoint not found",
			"path":  c.Path(),
		})
	})
}
