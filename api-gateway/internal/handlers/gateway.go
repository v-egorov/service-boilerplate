package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/services"
)

type GatewayHandler struct {
	registry *services.ServiceRegistry
	logger   *logrus.Logger
}

func NewGatewayHandler(registry *services.ServiceRegistry, logger *logrus.Logger) *GatewayHandler {
	return &GatewayHandler{
		registry: registry,
		logger:   logger,
	}
}

func (h *GatewayHandler) ProxyRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service URL
		serviceURL, err := h.registry.GetServiceURL(serviceName)
		if err != nil {
			h.logger.WithError(err).Error("Service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service unavailable"})
			return
		}

		// Parse service URL
		targetURL, err := url.Parse(serviceURL)
		if err != nil {
			h.logger.WithError(err).Error("Invalid service URL")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Modify the request
		c.Request.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.URL.Host = targetURL.Host

		// Add request ID to headers
		if requestID, exists := c.Get("request_id"); exists {
			c.Request.Header.Set("X-Request-ID", requestID.(string))
		}

		// Log the proxy request
		h.logger.WithFields(logrus.Fields{
			"service":    serviceName,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"request_id": c.GetString("request_id"),
		}).Info("Proxying request")

		// Custom director to handle request body properly
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Read request body
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					h.logger.WithError(err).Error("Failed to read request body")
					return
				}
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Custom error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			h.logger.WithError(err).Error("Proxy error")
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		}

		// Serve the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
