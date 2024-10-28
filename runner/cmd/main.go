package main

import (
	"os"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"streaming/runner/config"
	"streaming/runner/internal/runner"
)

func main() {
	natsURL := os.Getenv(config.EnvNatsURL)
	streamName := os.Getenv(config.EnvNatsStream)

	logger, err := zap.NewProduction(zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCaller())
	if err != nil {
		panic(err)
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	workers := config.DefaultWorkersCount

	workersCount := os.Getenv(config.EnvWorkersCount)
	if workersCount != "" {
		i64, err := strconv.ParseInt(workersCount, 10, 0)
		if err != nil {
			panic(err)
		}
		workers = i64
	}

	runner.New(js, logger, workers, streamName).Start()
}
