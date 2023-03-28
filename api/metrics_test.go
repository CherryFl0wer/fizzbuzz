package api

import (
	"FizzBuzz/domain"
	"FizzBuzz/service"
	mock_service "FizzBuzz/service/mock"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

type MetricsControllerSuite struct {
	suite.Suite
	logger *zap.Logger
	ctrl   *gomock.Controller
	Router *gin.Engine
	mms    *mock_service.MockMetricService
}

func (suite *MetricsControllerSuite) SetupSuite() {
	var err error
	suite.logger = zap.NewExample()
	suite.ctrl = gomock.NewController(suite.T())
	// Services needed
	suite.mms = mock_service.NewMockMetricService(suite.ctrl)
	suite.Router, err = Setup(nil, suite.mms, suite.logger)
	suite.Require().NoError(err)
}

func (suite *MetricsControllerSuite) TestMetricRequest() {
	tests := []struct {
		name           string
		expectedStatus int
		metric         *domain.MetricCountFizzBuzz
		serviceErr     error
		check          func(r *apitest.Response)
	}{
		{
			name:           "Ok 200",
			expectedStatus: http.StatusOK,
			metric: &domain.MetricCountFizzBuzz{
				Key:   "hash",
				Score: 2,
				Request: domain.FizzBuzzRequest{
					FstModulo: 3,
					SndModulo: 5,
					Limit:     10,
					FstStr:    "test",
					SndStr:    "test2",
				},
			},
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.Equal(`$.counter`, float64(2)))
				r.Assert(jsonpath.Equal(`$.request.limit`, float64(10)))
			},
		},
		{
			name:           "200 no counters",
			expectedStatus: http.StatusNoContent,
			metric:         nil,
			serviceErr:     service.ErrMetricsNoCountersFound,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.NotPresent(`$.counter`))
				r.Assert(jsonpath.NotPresent(`$.request`))
			},
		},
		{
			name:           "200 no data",
			expectedStatus: http.StatusNoContent,
			metric:         nil,
			serviceErr:     service.ErrMetricsNoDataFound,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.NotPresent(`$.counter`))
				r.Assert(jsonpath.NotPresent(`$.request`))
			},
		},

		{
			name:           "500 data is corrupted",
			expectedStatus: http.StatusInternalServerError,
			metric:         nil,
			serviceErr:     service.ErrMetricsNoRequestFound,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.NotPresent(`$.counter`))
				r.Assert(jsonpath.NotPresent(`$.request`))
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			if test.serviceErr != nil {
				suite.mms.EXPECT().MostRequested().Return(nil, test.serviceErr)
			} else {
				suite.mms.EXPECT().MostRequested().Return(test.metric, nil)
			}
			response := apitest.New().
				Debug().
				Handler(suite.Router).
				Get("/metrics").
				Expect(suite.T()).
				Status(test.expectedStatus)
			test.check(response)
			response.End()
		})
	}
}

func TestMetricsControllerSuite(t *testing.T) {
	suite.Run(t, new(MetricsControllerSuite))
}
