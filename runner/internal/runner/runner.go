package runner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/proto"

	"streaming/runner/internal/framer"
	"streaming/runner/internal/streaming/pb"
)

type Runner struct {
	js            jetstream.JetStream
	streamName    string
	frameSubject  string
	logger        *zap.Logger
	framerFactory map[pb.SourceType]func() framer.Framer
	sema          *semaphore.Weighted
	wg            *sync.WaitGroup
}

func (r *Runner) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	r.wg.Add(1)
	go r.start(ctx)

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		cancel()
	}()

	r.wg.Wait()

	r.logger.Info("gracefully stopped runner")
	os.Exit(0)
}

func (r *Runner) start(ctx context.Context) {
	defer r.wg.Done()

	jobConsumer, err := r.js.CreateConsumer(ctx, r.streamName, jetstream.ConsumerConfig{
		FilterSubject: "job",
	})
	if err != nil {
		return
	}

	messages, err := jobConsumer.Messages(jetstream.PullMaxMessages(1))
	if err != nil {
		r.logger.Error("push subscribe to messages failed", zap.Error(err))
		return
	}

	defer messages.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("stopping runner")
			return
		default:
			if r.sema.TryAcquire(1) {
				msg, err := messages.Next()
				if err != nil {
					if errors.Is(err, jetstream.ErrNoMessages) {
						time.Sleep(time.Second)
						continue
					}
				}

				var job pb.Job

				if err = proto.Unmarshal(msg.Data(), &job); err != nil {
					continue
				}

				go r.startFramer(ctx, &job)

				msg.Ack()
			} else {
				time.Sleep(time.Second)
			}
		}
	}
}

func (r *Runner) startFramer(ctx context.Context, job *pb.Job) {
	defer r.sema.Release(1)
	sourceFramer := r.framerFactory[job.Source.SourceType]()

	sourceFramer.Start(job.Source)

	for {
		frame, err := sourceFramer.Next()
		if err != nil {
			break
		}

		b, err := proto.Marshal(frame)
		if err != nil {
			break
		}

		_, err = r.js.PublishAsync(r.frameSubject, b)
		if err != nil {
			r.logger.Error("publish frame async", zap.Error(err))
			continue
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func New(js jetstream.JetStream, logger *zap.Logger, workers int64, streamName string) *Runner {
	return &Runner{
		js:            js,
		logger:        logger.Named("runner"),
		wg:            new(sync.WaitGroup),
		framerFactory: framer.Factory,
		sema:          semaphore.NewWeighted(workers),
		streamName:    streamName,
		frameSubject:  "frame",
	}
}
