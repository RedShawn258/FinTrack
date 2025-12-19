package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler exposes Prometheus metrics
// @Summary      Prometheus metrics
// @Description  Exposes Prometheus metrics for monitoring (outbox events, etc.)
// @Tags         monitoring
// @Produce      text/plain
// @Success      200  {string}  string  "Prometheus metrics"
// @Router       /metrics [get]
func MetricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
