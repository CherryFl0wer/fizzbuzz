package redis

import (
	"FizzBuzz"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: FizzBuzz.PrometheusNamespace,
		Subsystem: "redis",
		Name:      "error",
		Help:      "count error during request",
	})
	ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: FizzBuzz.PrometheusNamespace,
		Subsystem: "redis",
		Name:      "total_duration_seconds",
		Help:      "total duration of redis call",
		Buckets:   prometheus.ExponentialBuckets(.001, 1.5, 15),
	})
	ConnCreatedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: FizzBuzz.PrometheusNamespace,
		Subsystem: "redis",
		Name:      "conn_created",
		Help:      "count connection creation",
	}, []string{"service"})
	ConnClosedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: FizzBuzz.PrometheusNamespace,
		Subsystem: "redis",
		Name:      "conn_closed",
		Help:      "count connection closing",
	}, []string{"service"})
)

func KeyData(hash string) string {
	return fmt.Sprintf("fizzbuzz/data/%s", hash)
}

func KeyCounters() string {
	return "fizzbuzz/counters"
}

func NewRedis(host, port, pwd string) *redis.Client {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(host, port),
		Password: pwd,
		DB:       0,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			ConnCreatedCounter.WithLabelValues(cn.String()).Inc()
			return nil
		},
	})
	return redisCli
}

// RedisHealth will ping the client {every} time to see if it is healthy
func RedisHealth(cli *redis.Client, every time.Duration, logger *zap.Logger) {
	//nolint
	tick := time.Tick(every)
	name := cli.Conn().String()
	for range tick {
		status, err := cli.Ping(context.Background()).Result()
		if err != nil {
			if err = cli.Close(); err != nil {
				return
			}
			logger.Fatal("Can't ping redis",
				zap.String("redis-name", name))
			return
		}

		if strings.ToUpper(status) != "PONG" {
			ConnClosedCounter.WithLabelValues(name).Inc()
			if err = cli.Close(); err != nil {
				return
			}
			logger.Fatal("Something went wrong with redis connection",
				zap.String("redis-name", name))
		}
	}
}
