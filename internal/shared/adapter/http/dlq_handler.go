package http

import (
	"net/http"

	"github.com/ecelayes/pms-backend/internal/shared/adapter/redis"
	sharedContext "github.com/ecelayes/pms-backend/internal/shared/context"
	"github.com/labstack/echo/v4"
)

type DLQHandler struct {
	store *redis.DLQStore
}

func NewDLQHandler(store *redis.DLQStore) *DLQHandler {
	return &DLQHandler{store: store}
}

// GetStats returns DLQ statistics.
func (h *DLQHandler) GetStats(c echo.Context) error {
	ctx := sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c))

	stats, err := h.store.GetStats(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get DLQ length: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// GetMessages returns DLQ messages.
func (h *DLQHandler) GetMessages(c echo.Context) error {
	ctx := sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c))

	limit := int64(100)
	stats, err := h.store.GetMessages(ctx, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get DLQ messages: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stats)
}

// DeleteMessage removes a single message from the DLQ.
func (h *DLQHandler) DeleteMessage(c echo.Context) error {
	ctx := sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c))
	messageID := c.Param("id")

	deleted, err := h.store.DeleteMessage(ctx, messageID)
	if err != nil {
		if messageID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Message ID is required",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete message: " + err.Error(),
		})
	}

	if deleted == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Message not found in DLQ",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

// PurgeDLQ removes all messages from the DLQ.
func (h *DLQHandler) PurgeDLQ(c echo.Context) error {
	ctx := sharedContext.WithRequestID(c.Request().Context(), sharedContext.RequestIDFromEcho(c))

	deleted, err := h.store.Purge(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to purge DLQ: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}
