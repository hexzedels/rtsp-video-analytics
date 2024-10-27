package framer

import (
	"streaming/runner/internal/entity"
	"streaming/runner/internal/streaming/pb"
)

var Factory map[pb.SourceType]func() Framer = map[pb.SourceType]func() Framer{
	pb.SourceType_TYPE_VIDEO: NewVideoFramer,
}

type Framer interface {
	Start(*pb.Source) (*entity.SourceMetadata, error)
	Next() (*pb.Frame, error)
	Stop() error
}
