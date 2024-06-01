package client

import "streaming/api/internal/proto/pb"

type Client interface {
	pb.RunnerClient
}
