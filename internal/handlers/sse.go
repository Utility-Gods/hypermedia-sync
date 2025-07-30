package handlers

import (
	"fmt"
	"net/http"
	"time"

	"hypermedia-sync/internal/sse"

	"github.com/labstack/echo/v4"
)

func SSEHandler(hub *sse.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		originatorID := c.QueryParam("originator")
		if originatorID == "" {
			connCount := hub.GetOnlineCount()
			originatorID = fmt.Sprintf("sse-%d", connCount)
		}

		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering

		c.Response().WriteHeader(http.StatusOK)
		c.Response().Flush()

		fmt.Fprintf(c.Response().Writer, ": connected\n\n")
		c.Response().Flush()

		conn := &sse.Connection{
			ID:     originatorID,
			Writer: c.Response().Writer,
			Done:   make(chan struct{}),
		}

		hub.Register(conn)
		defer func() {
			hub.Unregister(conn)
			close(conn.Done)
		}()

		<-c.Request().Context().Done()
		return nil
	}
}

func HealthHandler(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "healthy",
		"service": "hypermedia-sync",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}