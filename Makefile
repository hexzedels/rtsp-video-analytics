# include .env

.PHONY: proto
proto:
	protoc --proto_path=./proto --go_out=runner proto/job.proto
	protoc --proto_path=./proto --go_out=api proto/api.proto proto/job.proto
	protoc --proto_path=./proto --go_out=orchestrator \
	--go-grpc_out=orchestrator proto/api.proto proto/job.proto