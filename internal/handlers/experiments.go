package handlers

import (
	"hypermedia-sync/internal/templates/pages"

	"github.com/labstack/echo/v4"
)

func ExperimentsListHandler(c echo.Context) error {
	experiments := []pages.Experiment{
		{
			ID:          "checkboxes",
			Name:        "10,000 Checkboxes",  
			Description: "Real-time synchronized checkboxes demonstrating hypermedia-driven state management with SSE",
			Path:        "/experiments/checkboxes",
			Status:      "Active",
		},
		{
			ID:          "canvas-draw-sync",
			Name:        "Canvas",  
			Description: "Collaborative real-time canvas where multiple users can draw, sketch, and create together using pure hypermedia",
			Path:        "/experiments/canvas-draw-sync",
			Status:      "Active",
		},
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		component := pages.ExperimentsListContent(experiments)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	component := pages.ExperimentsListPage(experiments)
	return component.Render(c.Request().Context(), c.Response().Writer)
}