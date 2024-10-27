package framer

import (
	"gocv.io/x/gocv"

	"streaming/runner/internal/entity"
	"streaming/runner/internal/streaming/pb"
)

type VideoFramer struct {
	jobID    string
	stream   *gocv.VideoCapture
	img      *gocv.Mat
	sequence int
}

func (r *VideoFramer) Start(source *pb.Source) (*entity.SourceMetadata, error) {
	video, err := gocv.OpenVideoCapture(source.Url)
	if err != nil {
		return nil, err
	}

	r.stream = video
	img := gocv.NewMat()
	r.img = &img

	return nil, nil
}

func (r *VideoFramer) Next() (*pb.Frame, error) {
	if ok := r.stream.Read(r.img); !ok {
		r.Stop()
		return nil, nil
	}

	r.sequence++

	return &pb.Frame{
		Id:        r.jobID,
		Sequence:  int32(r.sequence),
		Rows:      int32(r.img.Rows()),
		Cols:      int32(r.img.Cols()),
		FrameType: int32(r.img.Type()),
		Payload:   r.img.ToBytes(),
	}, nil
}

func (r *VideoFramer) Stop() error {
	if err := r.stream.Close(); err != nil {
		return err
	}

	if err := r.img.Close(); err != nil {
		return err
	}

	return nil
}

func NewVideoFramer() Framer {
	return &VideoFramer{}
}
