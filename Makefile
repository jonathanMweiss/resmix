all:
	protoc --go_out=plugins=grpc:. *.proto
	mv github.com/jonathanMweiss/resmix/mixnet.pb.go .
	rm -rf github.com
