package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/uncle3dev/velotrax-gateway-go/internal/grpc/client"
	authHandler "github.com/uncle3dev/velotrax-gateway-go/internal/handler/auth"
	"github.com/uncle3dev/velotrax-gateway-go/internal/middleware"
)

// Setup initializes all routes
func Setup(
	engine *gin.Engine,
	authClient *client.AuthClient,
	logger *zap.Logger,
	jwtSecret string,
) {
	// Apply global middleware
	engine.Use(middleware.Logger(logger))
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.CORS())

	// Auth handlers
	authH := authHandler.NewHandler(authClient, logger)

	// ── Public routes ──────────────────────────────────────────────────
	v1 := engine.Group("/v1")
	{
		// Auth endpoints (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		// Auth endpoints (protected - access token)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.RequireAuth(jwtSecret, logger))
		{
			authProtected.POST("/logout", authH.Logout)
		}

		// Token refresh (protected - refresh token)
		authRefresh := v1.Group("/auth")
		authRefresh.Use(middleware.RequireRefreshToken(jwtSecret, logger))
		{
			authRefresh.POST("/refresh", authH.RefreshToken)
		}
	}

	// ── Health check ───────────────────────────────────────────────────
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
