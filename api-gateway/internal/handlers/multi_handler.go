package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MultiHandler routes requests to different gin engines based on path prefix.
// MCP server routes (/mcp/*) bypass JWT validation while all other routes use the main router.
type MultiHandler struct {
	MainRouter  *gin.Engine
	HTTPHandler http.Handler
	Prefix      string
	Logger      *logrus.Logger
}

func (h *MultiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, h.Prefix) {
		h.Logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Debug("Routing to MCP server handler (no JWT)")
		h.HTTPHandler.ServeHTTP(w, r)
	} else {
		h.Logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Debug("Routing to main API handler")
		h.MainRouter.ServeHTTP(w, r)
	}
}
