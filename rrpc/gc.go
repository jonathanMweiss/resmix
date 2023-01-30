package rrpc

import (
	"github.com/jonathanMweiss/resmix/internal/syncmap"
	"time"
)

type GCable interface {
	GetStartTime() time.Time
	PrepareForDeletion()
}

func cleanmapAccordingToTTL[U comparable, T any](mp *syncmap.SyncMap[U, T], ttl time.Duration) {
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
