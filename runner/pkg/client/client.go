package client

import (
	"streaming/runner/internal/proto/pb"
)

type Client interface {
	pb.RunnerClient
}
