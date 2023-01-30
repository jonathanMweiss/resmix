package rrpc

import (
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
