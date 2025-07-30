package ultrathink

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"hypermedia-sync/internal/sse"
	"hypermedia-sync/internal/templates/experiments"

	"github.com/labstack/echo/v4"
)

var (
	canvas = experiments.CanvasState{
		Elements: []experiments.DrawingElement{},
		Width:    1200,
		Height:   800,
	}
	canvasMutex sync.RWMutex
)

func UltraThinkHandler(hub *sse.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		canvasMutex.RLock()
		defer canvasMutex.RUnlock()

		originatorID := fmt.Sprintf("canvas-%d-%d", time.Now().UnixNano(), rand.Intn(1000000))

		onlineCount := hub.GetOnlineCount()

		data := experiments.UltraThinkPageData{
			Canvas:       canvas,
			OriginatorID: originatorID,
			OnlineCount:  onlineCount,
		}

		if c.Request().Header.Get("HX-Request") == "true" {
			component := experiments.UltraThinkPageContent(data)
			return component.Render(c.Request().Context(), c.Response().Writer)
		}

		component := experiments.UltraThinkPageFull(data)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func DrawHandler(hub *sse.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		elementType := c.FormValue("type")
		elementData := c.FormValue("data")
		color := c.FormValue("color")
		brushSize := c.FormValue("brushSize")
		originatorID := c.Request().Header.Get("X-Originator-ID")

		if elementType == "" || elementData == "" {
			return c.String(400, "Missing drawing data")
		}

		element := experiments.DrawingElement{
			ID:        fmt.Sprintf("elem-%d-%d", time.Now().UnixNano(), rand.Intn(10000)),
			Type:      elementType,
			Data:      elementData,
			Color:     color,
			BrushSize: brushSize,
			User:      originatorID,
			Created:   time.Now(),
		}

		canvasMutex.Lock()
		canvas.Elements = append(canvas.Elements, element)
		canvasMutex.Unlock()

		var sseBuilder strings.Builder
		sseComponent := experiments.DrawingElementSSE(element)
		err := sseComponent.Render(c.Request().Context(), &sseBuilder)
		if err != nil {
			return c.String(500, "Error generating SSE HTML")
		}

		hub.Broadcast(sse.Event{
			Name:      "canvas-element-added",
			Data:      sseBuilder.String(),
			ExcludeID: originatorID,
		})

		var originatorBuilder strings.Builder
		originatorComponent := experiments.DrawingElementSSE(element)
		err = originatorComponent.Render(c.Request().Context(), &originatorBuilder)
		if err != nil {
			return c.String(500, "Error generating originator HTML")
		}
		
		return c.HTML(200, originatorBuilder.String())
	}
}

func ClearCanvasHandler(hub *sse.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		originatorID := c.Request().Header.Get("X-Originator-ID")

		canvasMutex.Lock()
		canvas.Elements = []experiments.DrawingElement{}
		canvasMutex.Unlock()

		var sseClearBuilder strings.Builder
		sseClearComponent := experiments.CanvasSVG(canvas)
		err := sseClearComponent.Render(c.Request().Context(), &sseClearBuilder)
		if err != nil {
			return c.String(500, "Error generating clear canvas SSE HTML")
		}

		hub.Broadcast(sse.Event{
			Name:      "canvas-cleared",
			Data:      sseClearBuilder.String(),
			ExcludeID: originatorID,
		})

		canvasMutex.RLock()
		data := experiments.UltraThinkPageData{
			Canvas:       canvas,
			OriginatorID: originatorID,
			OnlineCount:  0,
		}
		canvasMutex.RUnlock()

		var builder strings.Builder
		component := experiments.CanvasSVG(data.Canvas)
		err = component.Render(c.Request().Context(), &builder)
		if err != nil {
			return c.String(500, "Error generating canvas HTML")
		}

		return c.HTML(200, builder.String())
	}
}