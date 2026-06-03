package iam

import (
	"github.com/ecelayes/pms-backend/internal/iam/adapter"
	"github.com/ecelayes/pms-backend/internal/iam/adapter/http"
	"github.com/ecelayes/pms-backend/internal/iam/adapter/token"
	"github.com/ecelayes/pms-backend/internal/iam/application"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Module struct {
	AuthService *application.AuthService
	UserService *application.UserService
	OrgService  *application.OrganizationService
}

func NewModule(db *pgxpool.Pool, group *echo.Group, protected *echo.Group, emailService application.EmailService) *Module {
	userRepo := adapter.NewPostgresUserRepository(db)
	orgRepo := adapter.NewPostgresOrganizationRepository(db)
	authService := application.NewAuthService(userRepo, orgRepo, emailService)
	userService := application.NewUserService(userRepo, orgRepo)
	orgService := application.NewOrganizationService(orgRepo)
	protected.Use(token.Auth(authService))
	authH := http.NewAuthHandler(authService)
	userH := http.NewUserHandler(userService)
	orgH := http.NewOrganizationHandler(orgService)
	group.POST("/auth/login", authH.Login)
	group.POST("/auth/forgot-password", authH.ForgotPassword)
	group.POST("/auth/reset-password", authH.ResetPassword)
	protected.POST("/organizations", orgH.Create)
	protected.GET("/organizations", orgH.GetAll)
	protected.GET("/organizations/:id", orgH.GetByID)
	protected.PUT("/organizations/:id", orgH.Update)
	protected.DELETE("/organizations/:id", orgH.Delete)
	protected.POST("/users", userH.Create)
	protected.GET("/users", userH.GetAll)
	protected.GET("/users/:id", userH.GetByID)
	protected.PUT("/users/:id", userH.Update)
	protected.DELETE("/users/:id", userH.Delete)
	return &Module{
		AuthService: authService,
		UserService: userService,
		OrgService:  orgService,
	}
}
