genpb:
	protoc --proto_path=adapter/flowngine_adapter/pb adapter/flowngine_adapter/pb/*.proto --go_out=adapter/flowngine_adapter/pb --go_opt=paths=source_relative --go-grpc_out=adapter/flowngine_adapter/pb --go-grpc_opt=paths=source_relative
	
start:
	go run cmd/*.go start

.PHONY: genpb start