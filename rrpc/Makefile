all:
	protoc --go_out=plugins=grpc:. *.proto
	mv github.com/jonathanMweiss/resmixresmixrrpc.pb.go .
	rm -rf github.com
