@REM protoc --go_out=../.. ./xxxmail/mail.proto
@REM protoc --go_out=../.. ./xxxmail/mailver.proto
@REM protoc --go_out=../.. ./xxxmail/send.rpc.proto
@REM protoc --go_out=../.. --go-grpc_out=../.. ./xxxmail/send.grpc.proto

protoc --go_out=../.. ./user/user.proto

pause
