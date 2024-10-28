package inference

import (
	"context"
	"image"
	"sync"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"google.golang.org/protobuf/proto"

	"streaming/inference/internal/streaming/pb"
)

const (
	ratio   = 0.003921568627
	swapRGB = false
)

type Inference struct {
	net         *gocv.Net
	params      gocv.ImageToBlobParams
	outputNames []string

	wg             *sync.WaitGroup
	logger         *zap.Logger
	js             jetstream.JetStream
	streamName     string
	predictSubject string

	workersCount int
}

func New(logger *zap.Logger, js jetstream.JetStream, modelPath string, workers int, streamName string) *Inference {
	net := gocv.ReadNetFromONNX(modelPath)
	net.SetPreferableBackend(gocv.NetBackendOpenCV)
	net.SetPreferableTarget(gocv.NetTargetCPU)

	outputNames := getOutputNames(&net)
	if len(outputNames) == 0 {
		return nil
	}

	var (
		mean     = gocv.NewScalar(0, 0, 0, 0)
		padValue = gocv.NewScalar(144.0, 0, 0, 0)
	)

	params := gocv.NewImageToBlobParams(
		ratio,
		image.Pt(640, 640),
		mean,
		swapRGB,
		gocv.MatTypeCV32F,
		gocv.DataLayoutNCHW,
		gocv.PaddingModeLetterbox,
		padValue,
	)

	return &Inference{
		net:            &net,
		params:         params,
		outputNames:    outputNames,
		js:             js,
		wg:             new(sync.WaitGroup),
		logger:         logger.Named("inference"),
		workersCount:   workers,
		streamName:     streamName,
		predictSubject: "predict",
	}
}

func (r *Inference) Start() {
	ch := make(chan *pb.Frame)

	ctx := context.Background()

	for i := 0; i < r.workersCount; i++ {
		r.wg.Add(1)
		go r.worker(ch)
	}

	consumer, err := r.js.CreateConsumer(ctx, r.streamName, jetstream.ConsumerConfig{
		FilterSubject: "frame",
	})
	if err != nil {
		r.logger.Error("create frame consumer", zap.Error(err))
		return
	}

	frames, err := consumer.Messages()
	if err != nil {
		r.logger.Error("subscribe to messages", zap.Error(err))
		return
	}
	defer frames.Stop()

	for {
		msg, err := frames.Next()
		if err != nil {
			r.logger.Error("get next frame", zap.Error(err))
			break
		}

		msg.Ack()

		var frame pb.Frame

		if err := proto.Unmarshal(msg.Data(), &frame); err != nil {
			r.logger.Error("unmarshal frame", zap.Error(err))
			break
		}

		ch <- &frame
	}

	close(ch)

	r.wg.Wait()
}

func (r *Inference) worker(ch chan *pb.Frame) {
	defer r.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for frame := range ch {
		newImg, err := gocv.NewMatFromBytes(
			int(frame.Rows),
			int(frame.Cols),
			gocv.MatType(frame.FrameType),
			frame.Payload,
		)
		if err != nil {
			newImg.Close()
			return
		}

		if newImg.Empty() {
			newImg.Close()
			continue
		}

		predicts := r.detect(&newImg)

		// We send only frames with prediction
		if len(predicts) == 0 {
			continue
		}

		framePred := &pb.FramePrediction{
			Frame:    frame,
			Predicts: predicts,
		}

		b, err := proto.Marshal(framePred)
		if err != nil {
			newImg.Close()
			return
		}

		r.js.Publish(ctx, r.predictSubject, b)

		_ = b

		newImg.Close()
	}
}

func (r *Inference) detect(src *gocv.Mat) []*pb.Predict {
	blob := gocv.BlobFromImageWithParams(*src, r.params)
	defer blob.Close()

	// feed the blob into the detector
	r.net.SetInput(blob, "")

	// run a forward pass thru the network
	probs := r.net.ForwardLayers(r.outputNames)
	defer func() {
		for _, prob := range probs {
			prob.Close()
		}
	}()

	boxes, confidences, classIds := performDetection(probs)
	if len(boxes) == 0 {
		r.logger.Error("No classes detected")
		return nil
	}

	iboxes := r.params.BlobRectsToImageRects(boxes, image.Pt(src.Cols(), src.Rows()))
	indices := gocv.NMSBoxes(iboxes, confidences, scoreThreshold, nmsThreshold)

	return makeResult(iboxes, classes, classIds, indices)
}

// func MakeResult(img *gocv.Mat, boxes []image.Rectangle, classes []string, classIds []int, indices []int) []pb.Predict {
func makeResult(boxes []image.Rectangle, classes []string, classIds []int, indices []int) []*pb.Predict {
	var predicts []*pb.Predict

	for _, idx := range indices {
		if idx == 0 {
			continue
		}

		// gocv.Rectangle(img, image.Rect(boxes[idx].Min.X, boxes[idx].Min.Y, boxes[idx].Max.X, boxes[idx].Max.Y), color.RGBA{0, 255, 0, 0}, 2)
		// gocv.PutText(img, classes[classIds[idx]], image.Point{boxes[idx].Min.X, boxes[idx].Min.Y - 10}, gocv.FontHersheyPlain, 0.6, color.RGBA{0, 255, 0, 0}, 1)

		predicts = append(predicts, &pb.Predict{
			Class: classes[classIds[idx]],
			Rect: &pb.Rectangle{
				Min: &pb.Point{
					X: int32(boxes[idx].Min.X),
					Y: int32(boxes[idx].Min.Y),
				},
				Max: &pb.Point{
					X: int32(boxes[idx].Max.X),
					Y: int32(boxes[idx].Max.Y),
				},
			},
		})
	}

	return predicts
}

func performDetection(outs []gocv.Mat) ([]image.Rectangle, []float32, []int) {
	var classIds []int
	var confidences []float32
	var boxes []image.Rectangle

	// needed for yolov8
	gocv.TransposeND(outs[0], []int{0, 2, 1}, &outs[0])

	for _, out := range outs {
		out = out.Reshape(1, out.Size()[1])

		for i := 0; i < out.Rows(); i++ {
			cols := out.Cols()
			scoresCol := out.RowRange(i, i+1)

			scores := scoresCol.ColRange(4, cols)
			_, confidence, _, classIDPoint := gocv.MinMaxLoc(scores)

			if confidence > 0.5 {
				centerX := out.GetFloatAt(i, cols)
				centerY := out.GetFloatAt(i, cols+1)
				width := out.GetFloatAt(i, cols+2)
				height := out.GetFloatAt(i, cols+3)

				left := centerX - width/2
				top := centerY - height/2
				right := centerX + width/2
				bottom := centerY + height/2
				classIds = append(classIds, classIDPoint.X)
				confidences = append(confidences, float32(confidence))

				boxes = append(boxes, image.Rect(int(left), int(top), int(right), int(bottom)))
			}
		}
	}

	return boxes, confidences, classIds
}

func getOutputNames(net *gocv.Net) []string {
	var outputLayers []string
	for _, i := range net.GetUnconnectedOutLayers() {
		layer := net.GetLayer(i)
		layerName := layer.GetName()
		if layerName != "_input" {
			outputLayers = append(outputLayers, layerName)
		}
	}

	return outputLayers
}
