package rrpc

import (
	"github.com/jonathanMweiss/resmix/rrpc/internal"
	"time"
)

var (
	ttl = time.Second * 5
)

func init() {
	if !internal.InTest() {
		return
	}

	ttl = time.Second
}
