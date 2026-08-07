#! /bin/sh

# protoc --go_out=../.. ./xxxmail/mail.proto
# protoc --go_out=../.. ./xxxmail/mailver.proto
# protoc --go_out=../.. ./xxxmail/send.rpc.proto
# protoc --go_out=../.. --go-grpc_out=../.. ./xxxmail/send.grpc.proto

protoc --go_out=../.. ./user/user.proto
read -p "按任意键继续..."