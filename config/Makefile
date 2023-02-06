all:
	protoc --go_out=plugins=grpc:. *.proto
	mv github.com/jonathanMweiss/resmix/config/config.pb.go .
	rm -rf github.com
