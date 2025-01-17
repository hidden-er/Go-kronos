protoc --proto_path=/home/hiddener/Chamael/pkg/protobuf \
    --go_out=/home/hiddener/Chamael/pkg/protobuf --go_opt=paths=source_relative \
    --go-grpc_out=/home/hiddener/Chamael/pkg/protobuf --go-grpc_opt=paths=source_relative \
    Message.proto