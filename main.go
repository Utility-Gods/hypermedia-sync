package main

import (
	"fmt"
	"os"

	"hypermedia-sync/internal/experiments/checkboxes"
	"hypermedia-sync/internal/handlers"
	"hypermedia-sync/internal/sse"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize SSE hub
	hub := sse.NewHub()
	go hub.Run()

	e := echo.New()
	e.Use(middleware.Recover())
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

	// Start server on port from environment or 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port
	fmt.Printf("Server starting on %s\n", port)
	e.Logger.Fatal(e.Start(port))
}