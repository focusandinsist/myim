#! /bin/sh

protoc --go_out=../.. ./user/user.proto
protoc --go_out=../.. ./message/message.proto
protoc --go_out=../.. ./content/content.proto
protoc --go_out=../.. ./social/social.proto
read -p "按任意键继续..."
