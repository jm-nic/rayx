package controller

import (
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mhsanaei/3x-ui/v2/web/middleware"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"
	"golang.org/x/time/rate"

	"github.com/gin-gonic/gin"
)

// publicEndpointLimiter enforces a per-IP rate limit on the unauthenticated
// public API endpoints to mitigate brute-force and enumeration attempts.
// Allows a burst of 10 requests and then 1 request every 2 seconds (≈30 req/min).
var publicEndpointLimiter = middleware.NewIPRateLimiter(rate.Every(2*time.Second), 10)

// StopPublicRateLimiter stops the background cleanup goroutine of the package-level
// public-endpoint rate limiter. It should be called once during graceful shutdown.
func StopPublicRateLimiter() {
	publicEndpointLimiter.Stop()
}

// APIController handles the main API routes for the 3x-ui panel, including inbounds and server management.
type APIController struct {
	BaseController
	inboundController *InboundController
	serverController  *ServerController
	inboundService    service.InboundService
	Tgbot             service.Tgbot
}

// NewAPIController creates a new APIController instance and initializes its routes.
func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	return a
}

// checkAPIAuth is a middleware that returns 404 for unauthenticated API requests
// to hide the existence of API endpoints from unauthorized users
func (a *APIController) checkAPIAuth(c *gin.Context) {
	if !session.IsLogin(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
}

// initRouter sets up the API routes for inbounds, server, and other endpoints.
func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Public API group (no authentication required).
	// Rate limiting is applied here to protect against enumeration and brute-force.
	public := g.Group("/panel/api/public")
	public.Use(middleware.RateLimit(publicEndpointLimiter))
	public.GET("/client-expiry", a.getClientExpiryByUUID)
	public.GET("/client-expiry/:uuid", a.getClientExpiryByUUID)

	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkAPIAuth)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Extra routes
	api.GET("/backuptotgbot", a.BackuptoTgbot)
}

// getClientExpiryByUUID returns minimal expiration information for a single client UUID.
//
// Security rationale:
//   - Only a strict point-lookup by UUID is performed (WHERE id = ? LIMIT 1); there is
//     no pagination or filtering that could allow listing all registered UUIDs.
//   - UUID format is validated before hitting the database, but an invalid format returns
//     the same HTTP 404 as a valid-but-unknown UUID to prevent format-based oracles.
//   - The response only exposes expiry status (expiresAt, daysRemaining, expired) and
//     intentionally omits sensitive identity fields (email, remark, subId) and the raw
//     epoch timestamp to minimise information disclosure.
//   - The enclosing router group applies a per-IP rate limit (≈30 req/min) to make
//     brute-force enumeration of the 2^122 UUIDv4 space even less practical.
func (a *APIController) getClientExpiryByUUID(c *gin.Context) {
	rawUUID := strings.TrimSpace(c.Param("uuid"))
	if rawUUID == "" {
		rawUUID = strings.TrimSpace(c.Query("uuid"))
	}
	if rawUUID == "" {
		pureJsonMsg(c, http.StatusBadRequest, false, "uuid is required")
		return
	}

	// Validate UUID format before querying. We deliberately return 404 (not 400) for
	// malformed UUIDs so that callers cannot distinguish "bad format" from "not found"
	// and thus cannot use the error code as a format-validation oracle.
	if _, err := uuid.Parse(rawUUID); err != nil {
		pureJsonMsg(c, http.StatusNotFound, false, "not found")
		return
	}

	expiryTime, found, err := a.inboundService.GetClientExpiryByUUID(rawUUID)
	if err != nil {
		pureJsonMsg(c, http.StatusInternalServerError, false, "failed to query expiry time")
		return
	}
	if !found {
		pureJsonMsg(c, http.StatusNotFound, false, "not found")
		return
	}

	// Return only expiry-status fields. The input UUID and raw epoch milliseconds are
	// deliberately excluded to limit information exposure.
	response := gin.H{
		"expired": false,
	}
	if expiryTime > 0 {
		expiryDateTime := time.UnixMilli(expiryTime).UTC()
		response["expiresAt"] = expiryDateTime.Format("2006-01-02")

		remainingMs := expiryTime - time.Now().UnixMilli()
		response["daysRemaining"] = int64(math.Ceil(float64(remainingMs) / 86400000.0))
		response["expired"] = remainingMs <= 0
	}
	// When expiryTime == 0 the client never expires: expiresAt and daysRemaining are
	// omitted from the response (absent, not null) to keep the schema unambiguous.

	jsonObj(c, response, nil)
}

// BackuptoTgbot sends a backup of the panel data to Telegram bot admins.
func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}
