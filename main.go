package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"hypermedia-sync/internal/experiments/checkboxes"
	canvasdrawsync "hypermedia-sync/internal/experiments/canvas-draw-sync"
	"hypermedia-sync/internal/handlers"
	"hypermedia-sync/internal/sse"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func configureRateLimiter() echo.MiddlewareFunc {
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(10),  // 10 requests per second (higher for real-time app)
				Burst:     20,              // Allow bursts of up to 20 requests
				ExpiresIn: 1 * time.Minute, // Reset counters after 1 minute of inactivity
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			fmt.Printf("Rate limiter error: %v\n", err)
			return context.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Rate limit exceeded. Please slow down and try again.",
			})
		},
	}
	return middleware.RateLimiterWithConfig(config)
}

func main() {
	// Initialize SSE hub
	hub := sse.NewHub()
	go hub.Run()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(configureRateLimiter())
	e.Use(middleware.CORS())

	// Serve static files
	e.Static("/static", "static")

	// Main routes
	e.GET("/", handlers.ExperimentsListHandler)
	e.GET("/health", handlers.HealthHandler)
	e.GET("/events", handlers.SSEHandler(hub))

	// Experiment routes
	e.GET("/experiments/checkboxes", checkboxes.CheckboxesHandler(hub))
	e.POST("/experiments/checkboxes/toggle/:id", checkboxes.ToggleHandler(hub))
	
	e.GET("/experiments/canvas-draw-sync", canvasdrawsync.CanvasDrawSyncHandler(hub))
	e.POST("/experiments/canvas-draw-sync/draw", canvasdrawsync.DrawHandler(hub))
	e.POST("/experiments/canvas-draw-sync/clear", canvasdrawsync.ClearCanvasHandler(hub))

	// Start server on port from environment or 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port
	fmt.Printf("Server starting on %s\n", port)
	e.Logger.Fatal(e.Start(port))
}