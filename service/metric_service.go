package service

//go:generate ../.deps/mockgen -destination mock/metric_service.go -source metric_service.go

import (
	"FizzBuzz/domain"
	"FizzBuzz/repository"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

var (
	ErrMetricsNoCountersFound = errors.New("no metric for top request fizzbuzz was found")
	ErrMetricsNoRequestFound  = errors.New("no payload found from requested data")
	ErrMetricsNoDataFound     = errors.New("no data found from requested data")
)

type MetricService interface {
	Increment(request domain.ToBytes) error
	MostRequested() (*domain.MetricCountFizzBuzz, error)
}

type metricService struct {
	cacheRepo repository.CacheCounterRepository
	logger    *zap.Logger
}

func NewMetricService(cacheRepo repository.CacheCounterRepository,
	logger *zap.Logger) MetricService {
	return &metricService{
		logger:    logger,
		cacheRepo: cacheRepo,
	}
}

func (ms *metricService) Increment(request domain.ToBytes) error {
	ctx := context.Background()
	if err := ms.cacheRepo.IncrementRequest(ctx, request); err != nil {
		return err
	}

	return nil
}

func (ms *metricService) MostRequested() (*domain.MetricCountFizzBuzz, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	counter, err := ms.cacheRepo.GetCounters(ctx, -1, -1)
	if err != nil {
		ms.logger.Error("Failed to get top counter", zap.Error(err))
		return nil, err
	}

	if len(counter) == 0 {
		ms.logger.Debug("No counters")
		return nil, ErrMetricsNoCountersFound
	}

	mcfbr := domain.MetricCountFizzBuzz{
		Key:   counter[0].Key,
		Score: counter[0].ScoreCounter,
	}

	requestPayload, err := ms.cacheRepo.GetData(ctx, mcfbr.Key)
	if err != nil {
		ms.logger.Error("Failed to get data", zap.Error(err))
		return nil, ErrMetricsNoDataFound
	}
	fbr := domain.FromStrToRequestFB(requestPayload)
	if fbr == nil {
		ms.logger.Debug("No request")
		return nil, ErrMetricsNoRequestFound
	}

	mcfbr.Request = *fbr
	return &mcfbr, nil
}
