package main

import (
	"context"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

type dataServiceConfig struct {
	BaseURL  string
	Username string
	Password string
}

func main() {
	var (
		port      int
		logLevel  string
		cachePath string

		dataServiceConfig dataServiceConfig
	)

	pflag.IntVarP(&port, "port", "P", 8080, "port to run the server on")
	pflag.StringVarP(&logLevel, "log-level", "L", "", "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVar(&dataServiceConfig.BaseURL, "db-url", "", "database base url")
	pflag.StringVar(&dataServiceConfig.Username, "db-username", "", "database username")
	pflag.StringVar(&dataServiceConfig.Password, "db-password", "", "database password")
	pflag.StringVar(&cachePath, "cache-path", "", "path to file for config caching")
	pflag.Parse()

	// ctx for setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, log := logger(logLevel)
	defer log.Sync() // nolint:errcheck

	ds := dataService(ctx, dataServiceConfig)

	configs, err := ds.EventConfigs(ctx, "ITB-1006")
	if err != nil {
		log.Fatal("error", zap.Error(err))
	}

	log.Info("configs", zap.Any("", configs))
}
