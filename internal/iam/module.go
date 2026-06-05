package iam

import (
	"github.com/ecelayes/pms-backend/internal/iam/adapter"
	"github.com/ecelayes/pms-backend/internal/iam/adapter/http"
	"github.com/ecelayes/pms-backend/internal/iam/adapter/token"
	"github.com/ecelayes/pms-backend/internal/iam/application"
	"github.com/ecelayes/pms-backend/pkg/auth"
	"github.com/jackc/pgx/v5/pgxpool"
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
