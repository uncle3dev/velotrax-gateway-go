package order

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderpb "github.com/uncle3dev/velotrax-gateway-go/internal/gen/order"
	"github.com/uncle3dev/velotrax-gateway-go/internal/grpc/client"
)

type Handler struct {
	orderClient *client.OrderClient
	logger      *zap.Logger
}

type listOrdersBody struct {
	Page     int32  `json:"page"`
	PageSize int32  `json:"pageSize"`
	Status   string `json:"status"`
}

// NewHandler creates a new order handler.
func NewHandler(orderClient *client.OrderClient, logger *zap.Logger) *Handler {
	return &Handler{
		orderClient: orderClient,
		logger:      logger,
	}
}

// List handles GET/POST /v1/orders.
func (h *Handler) List(c *gin.Context) {
	page := int32(parseIntOrDefault(c.Query("page"), 1))
	pageSize := int32(parseIntOrDefault(c.Query("pageSize"), 20))
	status := c.Query("status")

	if c.Request.Method == http.MethodPost {
		var body listOrdersBody
		if err := c.ShouldBindJSON(&body); err == nil {
			if body.Page > 0 {
				page = body.Page
			}
			if body.PageSize > 0 {
				pageSize = body.PageSize
			}
			if body.Status != "" {
				status = body.Status
			}
		}
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	resp, err := h.orderClient.ListOrders(c.Request.Context(), &orderpb.ListOrdersRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	}, c.GetHeader("Authorization"))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Get handles GET /v1/orders/:id.
func (h *Handler) Get(c *gin.Context) {
	resp, err := h.orderClient.GetOrder(c.Request.Context(), &orderpb.GetOrderRequest{
		Id: c.Param("id"),
	}, c.GetHeader("Authorization"))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Tracking handles GET /v1/orders/:id/tracking.
func (h *Handler) Tracking(c *gin.Context) {
	resp, err := h.orderClient.GetOrderTracking(c.Request.Context(), &orderpb.GetOrderTrackingRequest{
		Id: c.Param("id"),
	}, c.GetHeader("Authorization"))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		h.logger.Error("unknown order service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	h.logger.Warn("order grpc error",
		zap.String("code", st.Code().String()),
		zap.String("message", st.Message()),
	)

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
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

func parseIntOrDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
