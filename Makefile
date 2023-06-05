clean:
	@rm -vf ./shared/proto/*.pb.go
protoc: clean
	@cd ./shared/proto && protoc --go_out=. --go-grpc_out=. *.proto