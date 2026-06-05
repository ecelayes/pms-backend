package iam

import (
	"github.com/ecelayes/pms-backend/internal/iam/adapter"
	httpAdapter "github.com/ecelayes/pms-backend/internal/iam/adapter/http"
	"github.com/ecelayes/pms-backend/internal/iam/adapter/token"
	"github.com/ecelayes/pms-backend/internal/iam/application"
	"github.com/ecelayes/pms-backend/pkg/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
	"github.com/labstack/echo/v4"
)

type Module struct {
	AuthService *application.AuthService
	UserService *application.UserService
	OrgService  *application.OrganizationService
}

// NewModule wires the IAM bounded context.
//
// SECURITY: The production wiring injects concrete crypto implementations:
//   - auth.NewBcryptPasswordHasher()      - bcrypt with DefaultCost
//   - auth.NewJWTTokenGenerator()          - JWT HS256 with explicit exp + alg confusion protection
//   - auth.NewCryptoRandomSaltGenerator()  - crypto/rand 256-bit salt
//
// There is NO way for external callers (HTTP, config, etc.) to swap these.
// The injection happens once at startup, in this file, with compile-time types.
func NewModule(db *pgxpool.Pool, group *echo.Group, protected *echo.Group, emailService application.EmailService) *Module {
	userRepo := adapter.NewPostgresUserRepository(db)
	orgRepo := adapter.NewPostgresOrganizationRepository(db)

	// SECURITY: Concrete crypto implementations - cannot be swapped at runtime
	passwordHasher := auth.NewBcryptPasswordHasher()
	tokenGenerator := auth.NewJWTTokenGenerator()
	saltGenerator := auth.NewCryptoRandomSaltGenerator()

	authService := application.NewAuthService(userRepo, orgRepo, emailService, passwordHasher, tokenGenerator, saltGenerator)
	userService := application.NewUserService(userRepo, orgRepo, passwordHasher, saltGenerator)
	orgService := application.NewOrganizationService(orgRepo)

	// Inject the same secure token generator into the middleware
	protected.Use(token.Auth(token.AuthConfig{
		SaltProvider:   authService,
		TokenGenerator: tokenGenerator,
	}))

	authH := httpAdapter.NewAuthHandler(authService)
	userH := httpAdapter.NewUserHandler(userService)
	orgH := httpAdapter.NewOrganizationHandler(orgService)

	// SECURITY: Rate limit auth endpoints to mitigate brute force / credential stuffing.
	// 5 requests/second with burst of 10. Per-IP.
	authLimiter := httpAdapter.NewIPRateLimiter(5, 10, 10*time.Minute)
	authGroup := group.Group("/auth", authLimiter.Middleware())

	authGroup.POST("/login", authH.Login)
	authGroup.POST("/forgot-password", authH.ForgotPassword)
	authGroup.POST("/reset-password", authH.ResetPassword)
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
