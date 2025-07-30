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
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		component := pages.ExperimentsListContent(experiments)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	component := pages.ExperimentsListPage(experiments)
	return component.Render(c.Request().Context(), c.Response().Writer)
}