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

	// decaying retry (seconds)
	// should go 2 -> 4 -> 8 -> 16 -> 30 -> 30 -> 30
	decayRetry := 0 * time.Second
	for {
		if decayRetry > 0 {
			log.Info("Waiting to restart handling updates", zap.Duration("duration", decayRetry))
			time.Sleep(decayRetry)
		}

		log.Debug("Building messenger")

		start := time.Now()
		m := newMessenger("")

		log.Info("Handling updates")

		if err := handleUpdates(m, ds); err != nil {
			log.Error("unable to handle updates", zap.Error(err))

			switch {
			case decayRetry == 0:
				decayRetry = 2 * time.Second
			case decayRetry >= 30*time.Second:
				decayRetry = 30 * time.Second
			default:
				decayRetry *= 2
			}
		}

		// restart the decaying retry
		if time.Since(start) > decayRetry {
			decayRetry = 2 * time.Second
		}

		m.Close()
	}
}
