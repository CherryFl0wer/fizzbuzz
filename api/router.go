package api

import (
	"FizzBuzz"
	"FizzBuzz/service"
	"path/filepath"
	"time"

	"os"

	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func Setup(fbService service.FizzBuzzService,
	metricService service.MetricService,
	logger *zap.Logger) (*gin.Engine, error) {
	router := gin.New()
	router.RemoveExtraSlash = true

	router.Use(ginzap.Ginzap(logger.Named("access"), time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	router.Use(MetricHttpRequest())

	router.GET("/", Index)
	router.GET("/prometheus-metrics", gin.WrapH(promhttp.Handler()))
	pprof.Register(router)

	{ // Exposed routes for users
		// Serv fizz buzz service
		SetupFizzBuzzAPI(fbService, metricService, router, logger)
		// Serv custom metrics
		SetupMetricsAPI(metricService, router, logger)
	}
	return router, nil
}

func Index(c *gin.Context) {
	exec, _ := os.Executable()
	c.JSON(200, gin.H{
		"version":   FizzBuzz.Version,
		"component": filepath.Base(exec),
	})
}
