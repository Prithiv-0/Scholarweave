package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Scholarweave/api/handlers"
	"Scholarweave/api/middleware"
	"Scholarweave/config"
	"Scholarweave/internal/database"
	"Scholarweave/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (optional, won't fail if missing)
	_ = godotenv.Load()

	cfg := config.Load()

	// Set up structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Connect to PostgreSQL
	if cfg.DatabaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := database.Migrate(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "ScholarWeave API v1",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			} else {
				slog.Error("unhandled error", "error", err, "path", c.Path(), "method", c.Method())
			}
			return c.Status(code).JSON(fiber.Map{
				"error":  message,
				"status": code,
				"time":   time.Now().Format(time.RFC3339),
				"path":   c.Path(),
				"method": c.Method(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.FrontendOrigin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
	app.Use(limiter.New(limiter.Config{
		Max:               60,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	// Routes
	setupRoutes(app, cfg, pool)

	// Serve frontend static files (for production / Render)
	app.Use("/", static.New("./frontend/dist", static.Config{
		Browse: false,
	}))

	// Handle 404 Not Found (after static files)
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Endpoint not found",
			"path":  c.Path(),
		})
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("shutting down server...")
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			slog.Error("server forced to shutdown", "error", err)
		}
	}()

	slog.Info("starting ScholarWeave API", "port", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func setupRoutes(app *fiber.App, cfg config.Config, pool *pgxpool.Pool) {
	// API v1 group
	v1 := app.Group("/api/v1")

	// Initialize handlers
	openAlexHandler := handlers.NewOpenAlexHandler(cfg.OpenAlexBaseURL, cfg.OpenAlexHTTPTimeout)
	graphService := services.NewGraphService(cfg.OpenAlexBaseURL, cfg.OpenAlexHTTPTimeout)
	graphHandler := handlers.NewGraphHandler(graphService)
	authService := services.NewAuthService(pool)
	libraryService := services.NewLibraryService(pool)
	authHandler := handlers.NewAuthHandler(authService, cfg.JWTSecret, cfg.JWTExpiration)
	libraryHandler := handlers.NewLibraryHandler(libraryService)

	// Root route
	app.Get("/api", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Welcome to ScholarWeave API",
			"endpoints": []string{
				"/api/v1/health",
				"/api/v1/auth/register",
				"/api/v1/auth/login",
				"/api/v1/users/me",
				"/api/v1/users/me/favorites",
				"/api/v1/users/me/lists",
				"/api/v1/papers/search?q=your_query",
				"/api/v1/papers/:id",
			},
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Health check with real status
	v1.Get("/health", func(c fiber.Ctx) error {
		dbStatus := "healthy"
		if err := pool.Ping(context.Background()); err != nil {
			dbStatus = "unhealthy"
		}

		openAlexStatus := "connected"
		if !openAlexHandler.PingUpstream() {
			openAlexStatus = "unreachable"
		}

		overall := "ok"
		if dbStatus != "healthy" || openAlexStatus != "connected" {
			overall = "degraded"
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    overall,
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
			"services": fiber.Map{
				"api":      "healthy",
				"database": dbStatus,
				"openalex": openAlexStatus,
			},
		})
	})

	// Authentication routes
	authGroup := v1.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)

	// User profile routes (auth required)
	usersGroup := v1.Group("/users", middleware.AuthRequired(cfg.JWTSecret))
	usersGroup.Get("/me", authHandler.Me)
	usersGroup.Get("/me/favorites", libraryHandler.ListFavorites)
	usersGroup.Post("/me/favorites", libraryHandler.AddFavorite)
	usersGroup.Delete("/me/favorites/:paperId", libraryHandler.RemoveFavorite)
	usersGroup.Get("/me/lists", libraryHandler.ListReadingLists)
	usersGroup.Post("/me/lists", libraryHandler.CreateReadingList)
	usersGroup.Delete("/me/lists/:listId", libraryHandler.DeleteReadingList)
	usersGroup.Post("/me/lists/:listId/items", libraryHandler.AddReadingListItem)
	usersGroup.Delete("/me/lists/:listId/items/:paperId", libraryHandler.RemoveReadingListItem)

	// Search papers
	v1.Get("/papers/search", func(c fiber.Ctx) error {
		if c.Query("q") == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Search query is required")
		}
		return openAlexHandler.SearchPapers(c)
	})

	// Get paper by ID
	v1.Get("/papers/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Paper ID is required")
		}
		return openAlexHandler.GetPaperByID(c)
	})

	// Get paper RDF
	v1.Get("/papers/:id/rdf", func(c fiber.Ctx) error {
		return openAlexHandler.GetPaperRDF(c)
	})

	// Knowledge graph
	v1.Get("/graph/paper/:id", graphHandler.GetPaperGraph)
	v1.Get("/graph/author/:id", graphHandler.GetAuthorGraph)
	v1.Get("/graph/concept/:id", graphHandler.GetConceptGraph)
	v1.Post("/graph/search", graphHandler.GetSearchGraph)
}
