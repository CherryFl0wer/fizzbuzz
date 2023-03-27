package main

import (
	"FizzBuzz"
	"FizzBuzz/api"
	"FizzBuzz/repository"
	fbRedis "FizzBuzz/repository/redis"
	"FizzBuzz/service"
	"context"
	goflag "flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

const prefix string = "fb"

type Config struct {
	Development bool   `mapstructure:"dev"`
	RedisHost   string `mapstructure:"redis-host"`
	RedisPort   string `mapstructure:"redis-port"`
	RedisPwd    string `mapstructure:"redis-pwd"`
	LogLevel    string `mapstructure:"log-level"`
	Listen      string `mapstructure:"listen"`
}

func GetConfig() (Config, error) {
	pflag.Bool("dev", false, "enable development mode")
	pflag.String("redis-host", "localhost", "host of redis")
	pflag.String("redis-port", "6379", "port for redis")
	pflag.String("redis-pwd", "", "redis password")
	pflag.String("log-level", "", "log level to use: debug, info, warn, error")
	pflag.String("listen", ":8080", "listen address")
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()

	var config Config
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return config, fmt.Errorf("impossible to parse flags, %w", err)
	}
	viper.SetEnvPrefix(prefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("unmarshalling of flag failed: %w", err)
	}

	return config, nil
}

func main() {
	config, err := GetConfig()
	if err != nil {
		fmt.Printf("Impossible to load data at startup: %s\n", err)
		os.Exit(1)
	}

	logger, err := initLog(config)
	if err != nil {
		fmt.Printf("Impossible to init logger: %s\n", err)
		os.Exit(1)
	}
	sugarLogger := logger.Sugar()

	logger.Info("Trying to connect to redis",
		zap.String("host", config.RedisHost),
		zap.String("port", config.RedisPort))
	redisCli := fbRedis.NewRedis(config.RedisHost, config.RedisPort, config.RedisPwd)
	if status, err := redisCli.Ping(context.Background()).Result(); err != nil || status != "PONG" {
		logger.Fatal("Impossible to connect to redis",
			zap.String("host", config.RedisHost),
			zap.String("port", config.RedisPort))
	}

	go fbRedis.RedisHealth(redisCli, 5*time.Second, logger)

	cacheRepo := repository.NewCacheCounterRepository(redisCli, logger)
	fbService := service.NewFizzBuzzService(logger)
	metricService := service.NewMetricService(cacheRepo, logger)

	router, err := api.Setup(fbService, metricService, logger)
	if err != nil {
		sugarLogger.Fatal(err)
	}

	logger.Fatal("fizzbuzz service crashed", zap.Error(router.Run(config.Listen)))
}

func initLog(config Config) (logger *zap.Logger, err error) {
	var zapConfig zap.Config
	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}
	if config.LogLevel != "" {
		var level zapcore.Level
		if err = level.Set(config.LogLevel); err != nil {
			return nil, err
		}
		zapConfig.Level.SetLevel(level)
	}

	logger, err = zapConfig.Build()
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.String("version", FizzBuzz.Version))
	return logger, nil
}
