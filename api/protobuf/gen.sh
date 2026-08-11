#! /bin/sh

protoc --go_out=../.. ./user/user.proto
protoc --go_out=../.. ./message/message.proto
read -p "按任意键继续..."
