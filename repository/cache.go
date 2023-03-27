package repository

//go:generate ../.deps/mockgen -destination mock/cache.go -source cache.go

import (
	"FizzBuzz/domain"
	"FizzBuzz/domain/usecase"
	fbRedis "FizzBuzz/repository/redis"
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrCacheKeyNotFound = errors.New("key requested not found")
	ErrMaxRetryTx       = errors.New("maximum retry reach for transaction")
)

type TxFunc = func(tx *redis.Tx) error

type CacheCounterRepository interface {
	IncrementRequest(ctx context.Context, request domain.ToBytes) error
	GetCounters(ctx context.Context, from, to int64) (domain.MetricCountersScores, error)
	GetData(ctx context.Context, key string) (string, error)
}

type cacheCounterRepository struct {
	maxRetry int
	client   *redis.Client
	logger   *zap.Logger
}

func NewCacheCounterRepository(redisCli *redis.Client, logger *zap.Logger) CacheCounterRepository {
	return &cacheCounterRepository{logger: logger, client: redisCli, maxRetry: 1000}
}

func (c *cacheCounterRepository) retryTx(ctx context.Context, tx TxFunc, keyObs string) error {
	// Retry if the key has been changed.
	for i := 0; i < c.maxRetry; i++ {
		err := c.client.Watch(ctx, tx, keyObs)
		if err == nil {
			c.logger.Debug("Done exec tx")
			return nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}

		fbRedis.ErrorCounter.Inc()
		return err
	}

	c.logger.Error("Try to execute transaction but failed after reaching max retry",
		zap.String("KeyObserved", keyObs))
	return ErrMaxRetryTx
}

func (c *cacheCounterRepository) IncrementRequest(ctx context.Context, request domain.ToBytes) error {
	data := request.ToBytes()
	hash, err := usecase.GetHash(data)
	if err != nil {
		return err
	}

	// Set if not exist data counter, and add counter to priorityQ
	tx := func(tx *redis.Tx) error {
		exist := tx.SetNX(ctx, fbRedis.KeyData(hash), data, 0)
		if exist.Err() != nil {
			fbRedis.ErrorCounter.Inc()
			return exist.Err()
		}

		c.logger.Debug("Redis metric", zap.Any("setnx", exist))

		var countersErr error
		if exist.Val() {
			c.logger.Debug("ZAdd metric")
			countersErr = tx.ZAdd(ctx, fbRedis.KeyCounters(), redis.Z{
				Score:  1,
				Member: hash,
			}).Err()
		} else {
			c.logger.Debug("ZIncrBy metric")
			countersErr = tx.ZIncrBy(ctx, fbRedis.KeyCounters(), 1, hash).Err()
		}

		if countersErr != nil {
			return countersErr
		}
		return nil
	}

	return c.retryTx(ctx, tx, fbRedis.KeyCounters())
}

func (c *cacheCounterRepository) GetCounters(ctx context.Context,
	from, to int64) (domain.MetricCountersScores, error) {
	scores := c.client.ZRangeWithScores(ctx, fbRedis.KeyCounters(), from, to)

	if scores.Err() != nil {
		fbRedis.ErrorCounter.Inc()
		return domain.MetricCountersScores{}, scores.Err()
	}

	return usecase.FromRedisZScoreToMetric(scores.Val()), nil
}

func (c *cacheCounterRepository) GetData(ctx context.Context,
	key string) (string, error) {
	res := c.client.Get(ctx, fbRedis.KeyData(key))
	if res.Err() != nil {
		fbRedis.ErrorCounter.Inc()
		return "", res.Err()
	}

	return res.Val(), nil
}
