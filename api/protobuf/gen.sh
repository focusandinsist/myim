#! /bin/sh

protoc --go_out=../.. ./user/user.proto
read -p "按任意键继续..."