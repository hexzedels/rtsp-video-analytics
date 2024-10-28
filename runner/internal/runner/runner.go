package runner

import (
	"context"
	"errors"
	"image"
	"image/color"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
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

				msg.Ack()

				var job pb.Job

				if err = proto.Unmarshal(msg.Data(), &job); err != nil {
					continue
				}

				go r.startFramer(ctx, &job)
				go r.startSaver(ctx)

				if err := os.Mkdir("./"+job.Id, 0777); err != nil {
					r.logger.Error("mkdir failed", zap.Error(err))
					return
				}
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
	defer sourceFramer.Stop()

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

func (r *Runner) startSaver(ctx context.Context) {
	consumer, err := r.js.CreateConsumer(ctx, r.streamName, jetstream.ConsumerConfig{
		FilterSubject: "predict",
	})
	if err != nil {
		return
	}

	predicts, err := consumer.Messages()
	if err != nil {
		r.logger.Error("subscribe to messages", zap.Error(err))
		return
	}

	for {
		p, err := predicts.Next()
		if err != nil {
			r.logger.Error("get next predict", zap.Error(err))
			break
		}

		p.Ack()

		var predict pb.FramePrediction

		if err := proto.Unmarshal(p.Data(), &predict); err != nil {
			r.logger.Error("unmarshal predict", zap.Error(err))
			break
		}

		img, err := gocv.NewMatFromBytes(
			int(predict.Frame.Rows),
			int(predict.Frame.Cols),
			gocv.MatType(predict.Frame.FrameType),
			predict.Frame.Payload,
		)
		if err != nil {
			img.Close()
			return
		}

		if img.Empty() {
			img.Close()
			continue
		}

		for _, pred := range predict.Predicts {
			gocv.Rectangle(
				&img,
				image.Rect(
					int(pred.Rect.Min.X),
					int(pred.Rect.Min.Y),
					int(pred.Rect.Max.X),
					int(pred.Rect.Max.Y),
				),
				color.RGBA{0, 255, 0, 0},
				2,
			)

			gocv.PutText(
				&img,
				pred.Class,
				image.Point{int(pred.Rect.Min.X), int(pred.Rect.Min.Y) - 10},
				gocv.FontHersheyPlain,
				0.6,
				color.RGBA{0, 255, 0, 0},
				1,
			)
		}

		fp := "./" + predict.Frame.Id + "/" + strconv.Itoa(int(predict.Frame.Sequence)) + ".jpg"
		ok := gocv.IMWrite(fp, img)
		r.logger.Info("write image to file", zap.String("filepath", fp), zap.Bool("ok", ok))
		img.Close()
		_ = img
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
