package pricing

import (
	"github.com/ecelayes/pms-backend/internal/pricing/adapter"
	"github.com/ecelayes/pms-backend/internal/pricing/adapter/http"
	"github.com/ecelayes/pms-backend/internal/pricing/application"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Module struct {
	Service *application.PricingService
}

func NewModule(db *pgxpool.Pool, protected *echo.Group) *Module {
	ratePlanRepo := adapter.NewPostgresRatePlanRepository(db)
	priceRuleRepo := adapter.NewPostgresPriceRuleRepository(db)
	service := application.NewPricingService(priceRuleRepo, ratePlanRepo)
	pricingH := http.NewPricingHandler(service)
	ratePlanH := http.NewRatePlanHandler(service)
	protected.POST("/pricing/bulk", pricingH.BulkUpdate)
	protected.GET("/pricing/rules", pricingH.GetRules)
	protected.DELETE("/pricing/rules/:id", pricingH.DeleteRule)
	protected.POST("/rate-plans", ratePlanH.Create)
	protected.GET("/rate-plans", ratePlanH.List)
	protected.GET("/rate-plans/:id", ratePlanH.GetByID)
	protected.PUT("/rate-plans/:id", ratePlanH.Update)
	protected.DELETE("/rate-plans/:id", ratePlanH.Delete)
	return &Module{Service: service}
}
