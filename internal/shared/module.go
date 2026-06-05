package shared

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/ecelayes/pms-backend/internal/shared/adapter/http"
	redisAdapter "github.com/ecelayes/pms-backend/internal/shared/adapter/redis"
)

type Module struct {
	DLQHandler *http.DLQHandler
}

func NewModule(rdb *redis.Client, admin *echo.Group, logger *zap.Logger) *Module {
	dlqStore := redisAdapter.NewDLQStore(rdb, logger)
	dlqHandler := http.NewDLQHandler(dlqStore)

	admin.GET("/dlq/stats", dlqHandler.GetStats)
	admin.GET("/dlq/messages", dlqHandler.GetMessages)
	admin.DELETE("/dlq/messages/:id", dlqHandler.DeleteMessage)
	admin.DELETE("/dlq/purge", dlqHandler.PurgeDLQ)

	return &Module{DLQHandler: dlqHandler}
}
