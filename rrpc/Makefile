all:
	protoc --go_out=plugins=grpc:. *.proto
	mv github.com/jonathanMweiss/resmix/rrpc/rrpc.pb.go .
	rm -rf github.com
