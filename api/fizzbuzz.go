package api

import (
	"FizzBuzz/domain"
	"FizzBuzz/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type fizzBuzzController struct {
	fbs    service.FizzBuzzService
	ms     service.MetricService
	logger *zap.Logger
}

type inputFizzBuzzRequest struct {
	domain.FizzBuzzRequest
}

func (i *inputFizzBuzzRequest) inputValidator() ValidationFormatter {
	return ValidationFormatter{
		structToJson: map[string]string{
			"FstModulo": "fst_mod",
			"SndModulo": "snd_mod",
			"Limit":     "limit",
			"FstStr":    "fst_str",
			"SndStr":    "snd_str",
		},
	}
}

func (fb *fizzBuzzController) Index(c *gin.Context) {
	var inp inputFizzBuzzRequest
	if err := c.ShouldBindJSON(&inp); err != nil {
		c.JSON(http.StatusBadRequest, BuildValidationError(err, inp.inputValidator()))
		return
	}

	if err := fb.ms.Increment(&inp); err != nil {
		fb.logger.Error("while incrementing request", zap.Error(err))
	}

	res := fb.fbs.SimpleFizzBuzz(inp.Limit, inp.FstModulo, inp.SndModulo, inp.FstStr, inp.SndStr)
	c.JSON(http.StatusOK, res)
}

func SetupFizzBuzzAPI(fbService service.FizzBuzzService,
	metricService service.MetricService,
	router *gin.Engine,
	logger *zap.Logger) {
	c := &fizzBuzzController{fbs: fbService, ms: metricService, logger: logger}
	router.POST("/fizzbuzz", c.Index)
}
