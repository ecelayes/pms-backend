package http

import (
	"net/http"
	
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type DLQHandler struct {
	client *redis.Client
}

func NewDLQHandler(client *redis.Client) *DLQHandler {
	return &DLQHandler{client: client}
}

type DLQMessage struct {
	OriginalID string `json:"original_id"`
	EventType  string `json:"event_type"`
	Payload    string `json:"payload"`
	FailedAt   int64  `json:"failed_at"`
	Retries    int    `json:"retries"`
}

type DLQStats struct {
	TotalMessages int64       `json:"total_messages"`
	Messages      []DLQMessage `json:"messages,omitempty"`
}

// GET /admin/dlq/stats - Get DLQ statistics
func (h *DLQHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()
	
	count, err := h.client.XLen(ctx, "reservation:events:dlq").Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get DLQ length: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, DLQStats{
		TotalMessages: count,
	})
}

// GET /admin/dlq/messages?limit=100 - Get DLQ messages
func (h *DLQHandler) GetMessages(c echo.Context) error {
	ctx := c.Request().Context()
	
	limit := int64(100)
	messages, err := h.client.XRange(ctx, "reservation:events:dlq", "-", "+").Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get DLQ messages: " + err.Error(),
		})
	}
	
	dlqMessages := make([]DLQMessage, 0, len(messages))
	for _, msg := range messages {
		if int64(len(dlqMessages)) >= limit {
			break
		}
		dlqMsg := DLQMessage{
			OriginalID: getStringValue(msg.Values, "original_id"),
			EventType:  getStringValue(msg.Values, "event_type"),
			Payload:    getStringValue(msg.Values, "payload"),
			FailedAt:   getInt64Value(msg.Values, "failed_at"),
			Retries:    getIntValue(msg.Values, "retries"),
		}
		dlqMessages = append(dlqMessages, dlqMsg)
	}
	
	return c.JSON(http.StatusOK, DLQStats{
		TotalMessages: int64(len(dlqMessages)),
		Messages:     dlqMessages,
	})
}

// DELETE /admin/dlq/messages/:id - Remove a message from DLQ
func (h *DLQHandler) DeleteMessage(c echo.Context) error {
	ctx := c.Request().Context()
	messageID := c.Param("id")
	
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message ID is required",
		})
	}
	
	deleted, err := h.client.XDel(ctx, "reservation:events:dlq", messageID).Result()
	if err != nil {
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

// DELETE /admin/dlq/purge - Purge all DLQ messages
func (h *DLQHandler) PurgeDLQ(c echo.Context) error {
	ctx := c.Request().Context()
	
	deleted, err := h.client.XDel(ctx, "reservation:events:dlq").Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to purge DLQ: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

func getStringValue(values map[string]interface{}, key string) string {
	if v, ok := values[key].(string); ok {
		return v
	}
	return ""
}

func getInt64Value(values map[string]interface{}, key string) int64 {
	if v, ok := values[key].(int64); ok {
		return v
	}
	if v, ok := values[key].(string); ok {
		var i int64
		for _, c := range v {
			if c >= '0' && c <= '9' {
				i = i*10 + int64(c-'0')
			}
		}
		return i
	}
	return 0
}

func getIntValue(values map[string]interface{}, key string) int {
	if v, ok := values[key].(int); ok {
		return v
	}
	if v, ok := values[key].(int64); ok {
		return int(v)
	}
	if v, ok := values[key].(string); ok {
		var i int
		for _, c := range v {
			if c >= '0' && c <= '9' {
				i = i*10 + int(c-'0')
			}
		}
		return i
	}
	return 0
}
