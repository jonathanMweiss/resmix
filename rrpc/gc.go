package rrpc

import (
	"context"
	"github.com/jonathanMweiss/resmix/internal/msync"
	"time"
)

type GCable interface {
	GetStartTime() time.Time
	PrepareForDeletion()
}

// not using generics because something might not be a GCable, and i still want it to be deleted.
func cleanmapAccordingToTTL[U comparable, T any](mp *msync.Map[U, T], ttl time.Duration) {
	mp.Map.Range(func(key, value interface{}) bool {
		v, ok := value.(GCable)
		if !ok {
			mp.Map.Delete(key)
			return true
		}

		if time.Since(v.GetStartTime()) < ttl {
			return true
		}

		v.PrepareForDeletion()
		mp.Map.Delete(key)

		return true
	})
}

const ttl = time.Second * 5

func foreverCleanup[U comparable, T any](ctx context.Context, mp *msync.Map[U, T]) {
	dropFromMap := time.NewTicker(ttl)
	for {
		select {
		case <-dropFromMap.C:
			cleanmapAccordingToTTL(mp, ttl)
		case <-ctx.Done():
			return
		}
	}
}
