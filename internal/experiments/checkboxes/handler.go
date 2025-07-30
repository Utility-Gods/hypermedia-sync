package checkboxes

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"hypermedia-sync/internal/sse" 
	"hypermedia-sync/internal/templates/experiments"

	"github.com/labstack/echo/v4"
)

var (
	checkboxes = make(map[int]bool)
	mu         sync.RWMutex
)

func init() {
	for i := 1; i <= 10000; i++ {
		checkboxes[i] = false
	}
}

func CheckboxesHandler(hub *sse.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		mu.RLock()
		defer mu.RUnlock()

		var cbData []experiments.CheckboxData

		for i := 1; i <= 10000; i++ {
			checked := checkboxes[i]
			cbData = append(cbData, experiments.CheckboxData{ID: i, Checked: checked})
		}

		// Generate unique originator ID for this page load
		originatorID := fmt.Sprintf("page-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))

		// Get current online count
		onlineCount := hub.GetOnlineCount()

		data := experiments.CheckboxPageData{
			Checkboxes:   cbData,
			OriginatorID: originatorID,
			OnlineCount:  onlineCount,
		}

		// Dual response pattern - check if it's an HTMX request
		if c.Request().Header.Get("HX-Request") == "true" {
			// Return just the content for HTMX requests
			component := experiments.CheckboxesPageContent(data)
			return component.Render(c.Request().Context(), c.Response().Writer)
		}

		// Return full page for direct access
		component := experiments.CheckboxesPageFull(data)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func ToggleHandler(hub *sse.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "Invalid checkbox ID")
		}

		if id < 1 || id > 10000 {
			return c.String(400, "Checkbox ID out of range")
		}

		// Get originator ID
		originatorID := c.Request().Header.Get("X-Originator-ID")

		// Toggle checkbox state
		mu.Lock()
		checkboxes[id] = !checkboxes[id]
		newState := checkboxes[id]
		mu.Unlock()

		// Generate HTML for this checkbox
		cb := experiments.CheckboxData{ID: id, Checked: newState}
		
		// Generate HTML for SSE broadcast (excluding originator)
		var sseBuilder strings.Builder
		sseComponent := experiments.CheckboxItemSSEComplete(cb)
		err = sseComponent.Render(c.Request().Context(), &sseBuilder)
		if err != nil {
			return c.String(500, "Error generating SSE HTML")
		}

		// Broadcast only the affected checkbox (excluding originator)
		hub.Broadcast(sse.Event{
			Name:      fmt.Sprintf("checkbox-%d-updated", id),
			Data:      sseBuilder.String(),
			ExcludeID: originatorID,
		})

		// Return updated HTML to originator for immediate feedback
		var originatorBuilder strings.Builder
		originatorComponent := experiments.CheckboxItemSSEComplete(cb)
		err = originatorComponent.Render(c.Request().Context(), &originatorBuilder)
		if err != nil {
			return c.String(500, "Error generating originator HTML")
		}
		
		return c.HTML(200, originatorBuilder.String())
	}
}
