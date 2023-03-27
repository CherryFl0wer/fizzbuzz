package repository

import (
	"FizzBuzz/domain"
	"FizzBuzz/domain/usecase"
	fbRedis "FizzBuzz/repository/redis"
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type CacheCounterRepositorySuite struct {
	suite.Suite
	redisServer *miniredis.Miniredis
	redisClient *redis.Client
	logger      *zap.Logger
	ccRepo      CacheCounterRepository
}

func (suite *CacheCounterRepositorySuite) SetupTest() {
	var err error
	suite.redisServer, err = miniredis.Run()
	suite.Require().NoError(err)

	host := strings.Split(suite.redisServer.Addr(), ":")
	suite.redisClient = fbRedis.NewRedis(host[0], host[1], "")
	suite.Require().NotNil(suite.redisClient)

	suite.Require().NoError(suite.redisClient.Ping(context.Background()).Err())

	suite.logger = zap.NewExample()
	suite.ccRepo = NewCacheCounterRepository(suite.redisClient, suite.logger)

}

func (suite *CacheCounterRepositorySuite) cleanRedis(testName string) {
	suite.logger.Info("Finished, flushing redis", zap.String("Test name", testName))
	suite.redisClient.FlushDB(context.Background())
}

func (suite *CacheCounterRepositorySuite) TestIncrementRequest() {
	tests := []struct {
		name      string
		nbRequest int
		request   domain.ToBytes
	}{
		{
			name:      "Simple Increment",
			nbRequest: 1,
			request: &domain.FizzBuzzRequest{
				FstModulo: 3,
				SndModulo: 5,
				Limit:     10,
				FstStr:    "three",
				SndStr:    "five",
			},
		},
		{
			name:      "Increment three times",
			nbRequest: 3,
			request: &domain.FizzBuzzRequest{
				FstModulo: 4,
				SndModulo: 5,
				Limit:     10,
				FstStr:    "four",
				SndStr:    "five",
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			for i := 0; i < test.nbRequest; i++ {
				err := suite.ccRepo.IncrementRequest(context.Background(), test.request)
				suite.Require().NoError(err)
			}

			hash, err := usecase.GetHash(test.request.ToBytes())
			suite.Require().NoError(err)
			val, err := suite.redisServer.ZScore(fbRedis.KeyCounters(), hash)
			suite.Require().NoError(err)
			suite.EqualValues(test.nbRequest, val)

			suite.cleanRedis(test.name)
		})

	}
}

func (suite *CacheCounterRepositorySuite) TestValidData() {
	tests := []struct {
		name    string
		request domain.ToBytes
	}{
		{
			name: "Get payload and check validity",
			request: &domain.FizzBuzzRequest{
				FstModulo: 3,
				SndModulo: 5,
				Limit:     10,
				FstStr:    "three",
				SndStr:    "five",
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// Should not call increment request
			err := suite.ccRepo.IncrementRequest(context.Background(), test.request)
			suite.Require().NoError(err)

			hash, err := usecase.GetHash(test.request.ToBytes())
			suite.Require().NoError(err)

			payload, err := suite.ccRepo.GetData(context.Background(), hash)
			suite.Require().NoError(err)
			suite.Require().NotEmpty(payload)

			fbr := test.request.(*domain.FizzBuzzRequest)
			parsedRequest := domain.FromStrToRequestFB(payload)

			suite.EqualValues(fbr.FstStr, parsedRequest.FstStr)
			suite.EqualValues(fbr.SndStr, parsedRequest.SndStr)
			suite.EqualValues(fbr.FstModulo, parsedRequest.FstModulo)
			suite.EqualValues(fbr.SndModulo, parsedRequest.SndModulo)
			suite.EqualValues(fbr.Limit, parsedRequest.Limit)

			suite.cleanRedis(test.name)
		})
	}
}

func (suite *CacheCounterRepositorySuite) TestCounters() {
	tests := []struct {
		name             string
		from             int64
		to               int64
		requests         []domain.ToBytes
		expectedCounters int
		ranking          []int
	}{
		{
			name: "Top 2",
			from: 0,
			to:   1,
			requests: []domain.ToBytes{
				&domain.FizzBuzzRequest{
					FstModulo: 3,
					SndModulo: 5,
					Limit:     10,
					FstStr:    "three",
					SndStr:    "five",
				},
				&domain.FizzBuzzRequest{
					FstModulo: 3,
					SndModulo: 5,
					Limit:     10,
					FstStr:    "three",
					SndStr:    "five",
				},
				&domain.FizzBuzzRequest{
					FstModulo: 2,
					SndModulo: 4,
					Limit:     8,
					FstStr:    "two",
					SndStr:    "four",
				},
			},
			expectedCounters: 2,
			ranking:          []int{2, 1},
		},
		{
			name: "Top 1",
			from: -1,
			to:   -1,
			requests: []domain.ToBytes{
				&domain.FizzBuzzRequest{
					FstModulo: 3,
					SndModulo: 5,
					Limit:     10,
					FstStr:    "three",
					SndStr:    "five",
				},
				&domain.FizzBuzzRequest{
					FstModulo: 3,
					SndModulo: 5,
					Limit:     10,
					FstStr:    "three",
					SndStr:    "five",
				},
				&domain.FizzBuzzRequest{
					FstModulo: 2,
					SndModulo: 4,
					Limit:     8,
					FstStr:    "two",
					SndStr:    "four",
				},
			},
			expectedCounters: 1,
			ranking:          []int{2},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			for i := 0; i < len(test.requests); i++ {
				err := suite.ccRepo.IncrementRequest(context.Background(), test.requests[i])
				suite.Require().NoError(err)
			}

			counters, err := suite.ccRepo.GetCounters(context.Background(), test.from, test.to)
			suite.Require().NoError(err)

			suite.Len(counters, test.expectedCounters)
			for i := 0; i < test.expectedCounters; i++ {
				suite.EqualValues(test.ranking[len(test.ranking)-i-1], counters[i].ScoreCounter)
			}
			suite.cleanRedis(test.name)
		})
	}
}

func TestCacheCounterRepositorySuite(t *testing.T) {
	suite.Run(t, new(CacheCounterRepositorySuite))
}
