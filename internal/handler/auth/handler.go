package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	auth "github.com/uncle3dev/velotrax-gateway-go/internal/gen/auth"
	"github.com/uncle3dev/velotrax-gateway-go/internal/grpc/client"
	"github.com/uncle3dev/velotrax-gateway-go/internal/middleware"
)

type Handler struct {
	authClient *client.AuthClient
	logger     *zap.Logger
	validator  *validator.Validate
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"fullName" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LogoutRequest struct {
	// No body needed, userID comes from token
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// NewHandler creates a new Auth handler
func NewHandler(authClient *client.AuthClient, logger *zap.Logger) *Handler {
	return &Handler{
		authClient: authClient,
		logger:     logger,
		validator:  validator.New(),
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	h.logger.Info("register request", zap.Any("request", req))
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn("validation error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation error"})
		return
	}

	grpcReq := &auth.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	}

	resp, err := h.authClient.Register(context.Background(), grpcReq)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": resp.UserId,
		"email":   resp.Email,
	})
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn("validation error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation error"})
		return
	}

	grpcReq := &auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authClient.Login(context.Background(), grpcReq)
	if err != nil {
		h.respondError(c, err)
		return
	}

	// Log successful login
	h.logger.Info("Successful user login",
		zap.String("access_token", resp.AccessToken),
		zap.String("refresh_token", resp.RefreshToken),
		zap.Int64("expires_in", resp.ExpiresIn),
		zap.String("user_id", resp.User.Id),
		zap.String("email", resp.User.Email),
		zap.Strings("roles", resp.User.Roles),
	)

	// ✅ Trả về struct LoginResponse
	c.JSON(http.StatusOK, &auth.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn, // Chuyển int64 -> int
		User:         resp.User,      // Trả về UserDetail nếu có
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &auth.LogoutRequest{
		UserId: userID,
	}

	resp, err := h.authClient.Logout(context.Background(), grpcReq)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid refresh request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn("validation error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation error"})
		return
	}

	grpcReq := &auth.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.authClient.RefreshToken(context.Background(), grpcReq)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": resp.AccessToken,
		"expires_in":   resp.ExpiresIn,
	})
}

// respondError maps gRPC status codes to HTTP status codes
func (h *Handler) respondError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		h.logger.Error("unknown error type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	h.logger.Warn("grpc error",
		zap.String("code", st.Code().String()),
		zap.String("message", st.Message()),
	)

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument"})
	case codes.Unavailable, codes.DeadlineExceeded:
		c.JSON(http.StatusBadGateway, gin.H{"error": "service unavailable"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
