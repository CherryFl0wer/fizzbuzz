package usecase

import (
	"FizzBuzz/domain"
	"errors"
	"fmt"

	"github.com/cespare/xxhash/v2"
	"github.com/redis/go-redis/v9"
)

var (
	ErrHashingFailed = errors.New("hashing failed")
)

func GetHash(data []byte) (string, error) {
	hasher := xxhash.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", ErrHashingFailed
	}
	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

func FromRedisZScoreToMetric(scores []redis.Z) domain.MetricCountersScores {
	mcs := make(domain.MetricCountersScores, len(scores))
	for i, score := range scores {
		mcs[i] = domain.MetricCounterScore{
			Key:          score.Member.(string),
			ScoreCounter: int(score.Score),
		}
	}

	return mcs
}
