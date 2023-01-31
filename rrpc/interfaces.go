package rrpc

type ClientConn interface {
	DirectCall(rq *Request) error
	AsyncDirectCall(rq *Request) (<-chan error, error)
	RobustCall(rq *Request) error
}
