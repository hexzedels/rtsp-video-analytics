package main

import (
	"os"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"streaming/inference/config"
	"streaming/inference/internal/inference"
)

func main() {
	natsURL := os.Getenv(config.EnvNatsURL)
	modelPath := os.Getenv(config.EnvModelPath)

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
		workers, err = strconv.Atoi(workersCount)
		if err != nil {
			panic(err)
		}
	}

	inference.New(logger, js, modelPath, workers).Start()
}
