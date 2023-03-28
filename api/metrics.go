package api

import (
	"FizzBuzz/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type metricsController struct {
	ms     service.MetricService
	logger *zap.Logger
}

func ParseMetricsError(err error) (int, ErrorResponse) {
	if errors.Is(err, service.ErrMetricsNoCountersFound) ||
		errors.Is(err, service.ErrMetricsNoDataFound) {
		return http.StatusNoContent, ErrorResponse{
			Message: "No data was found for top metric",
		}
	} else if errors.Is(err, service.ErrMetricsNoRequestFound) {
		return http.StatusInternalServerError, ErrorResponse{
			Message: "Data has been corrupted",
		}
	}

	return http.StatusInternalServerError, ErrorResponse{Message: "Sorry something went wrong"}
}

func SetupMetricsAPI(ms service.MetricService, router *gin.Engine, logger *zap.Logger) {
	mc := &metricsController{
		logger: logger,
		ms:     ms,
	}
	router.GET("/metrics", mc.Index)
}

func (mc *metricsController) Index(ctx *gin.Context) {
	res, err := mc.ms.MostRequested()
	if err != nil {
		code, errResp := ParseMetricsError(err)
		ctx.JSON(code, errResp)
		return
	}

	ctx.JSON(http.StatusOK, res)
}
