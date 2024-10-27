package runner

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Runner struct {
	js     jetstream.JetStream
	logger *zap.Logger
}

func (r *Runner) Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	r.logger.Info("gracefully stopped server")
	os.Exit(0)
}

func New(js jetstream.JetStream, logger *zap.Logger) *Runner {
	return &Runner{
		js:     js,
		logger: logger.Named("runner"),
	}
}
