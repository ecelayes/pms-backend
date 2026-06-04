package shared

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/ecelayes/pms-backend/internal/shared/adapter/http"
)

type Module struct {
	DLQHandler *http.DLQHandler
}

func NewModule(rdb *redis.Client, admin *echo.Group) *Module {
	dlqHandler := http.NewDLQHandler(rdb)
	
	admin.GET("/dlq/stats", dlqHandler.GetStats)
	admin.GET("/dlq/messages", dlqHandler.GetMessages)
	admin.DELETE("/dlq/messages/:id", dlqHandler.DeleteMessage)
	admin.DELETE("/dlq/purge", dlqHandler.PurgeDLQ)
	
	return &Module{DLQHandler: dlqHandler}
}
