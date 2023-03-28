package api

import (
	mock_repository "FizzBuzz/repository/mock"
	"FizzBuzz/service"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

type FizzBuzzControllerSuite struct {
	suite.Suite
	logger        *zap.Logger
	ctrl          *gomock.Controller
	Router        *gin.Engine
	mockCacheRepo *mock_repository.MockCacheCounterRepository
	fbs           service.FizzBuzzService
	ms            service.MetricService
}

func (suite *FizzBuzzControllerSuite) SetupSuite() {
	var err error
	suite.logger = zap.NewExample()
	suite.ctrl = gomock.NewController(suite.T())
	// Repo needed
	suite.mockCacheRepo = mock_repository.NewMockCacheCounterRepository(suite.ctrl)
	// Services needed
	suite.fbs = service.NewFizzBuzzService(suite.logger)
	suite.ms = service.NewMetricService(suite.mockCacheRepo, suite.logger)
	suite.Router, err = Setup(suite.fbs, suite.ms, suite.logger)
	suite.Require().NoError(err)
}

func (suite *FizzBuzzControllerSuite) TestFizzbuzzJsonRequest() {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
		check          func(r *apitest.Response)
	}{
		{
			name: "Ok 200",
			body: `{
	"fst_mod": 3,
	"snd_mod": 5,
	"limit": 15,
	"fst_str": "test",
	"snd_str": "test2"
}`,
			expectedStatus: http.StatusOK,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.Len(`$`, 15))
				r.Assert(jsonpath.Equal(`$[14]`, "testtest2"))
			},
		},
		{
			name: "Error required fields",
			body: `{
	"fst_mod": 3,
	"limit": 15,
	"fst_str": "test",
	"snd_str": "test2"
}`,
			expectedStatus: http.StatusBadRequest,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.Present(`$.errors`))
				r.Assert(jsonpath.Equal(`$.errors[0].field_name`, "snd_mod"))
			},
		},
		{
			name: "Error multiple required fields",
			body: `{
	"limit": 15,
	"fst_str": "test",
	"snd_str": "test2"
}`,
			expectedStatus: http.StatusBadRequest,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.Present(`$.errors`))
				r.Assert(jsonpath.Equal(`$.errors[0].field_name`, "fst_mod"))
				r.Assert(jsonpath.Equal(`$.errors[1].field_name`, "snd_mod"))
			},
		},
		{
			name: "Error limit fields",
			body: `{
    "fst_mod": -1,
	"snd_mod": 5,
	"limit": 15,
	"fst_str": "test",
	"snd_str": "test2"
}`,
			expectedStatus: http.StatusBadRequest,
			check: func(r *apitest.Response) {
				r.Assert(jsonpath.Present(`$.errors`))
				r.Assert(jsonpath.Equal(`$.errors[0].field_name`, "fst_mod"))
			},
		},
	}

	suite.mockCacheRepo.EXPECT().IncrementRequest(gomock.Any(), gomock.Any()).MaxTimes(len(tests))
	for _, test := range tests {
		suite.Run(test.name, func() {
			response := apitest.New().
				Debug().
				Handler(suite.Router).
				Post("/fizzbuzz").
				Body(test.body).
				Expect(suite.T()).
				Status(test.expectedStatus)
			test.check(response)
			response.End()
		})
	}
}

func TestFizzBuzzControllerSuite(t *testing.T) {
	suite.Run(t, new(FizzBuzzControllerSuite))
}
