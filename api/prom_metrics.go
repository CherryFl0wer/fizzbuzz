package api

import (
	"FizzBuzz"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: FizzBuzz.PrometheusNamespace,
		Subsystem: "http",
		Name:      "inflight",
		Help:      "numbers of cancelled http call",
	})
	HttpDurations = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: FizzBuzz.PrometheusNamespace,
		Subsystem: "http",
		Name:      "total_duration_seconds",
		Help:      "total duration of the handler",
		Buckets:   prometheus.ExponentialBuckets(.001, 1.5, 15),
	},
		[]string{"handler", "method", "code"},
	)
)

func MetricHttpRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		begin := time.Now()
		HttpInFlight.Inc()
		c.Next()
		HttpInFlight.Dec()
		observer := HttpDurations.With(prometheus.Labels{
			"handler": c.FullPath(),
			"method":  c.Request.Method,
			"code":    strconv.Itoa(c.Writer.Status()),
		})
		observer.Observe(time.Since(begin).Seconds())
	}
}
