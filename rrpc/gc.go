package rrpc

import (
	"sync"
	"time"
)

type GCable interface {
	GetStartTime() time.Time
	PrepareForDeletion()
}

func cleanmapAccordingToTTL(mp *sync.Map, ttl time.Duration) {
	mp.Range(func(key, value interface{}) bool {
		v, ok := value.(GCable)
		if !ok {
			mp.Delete(key)
			return true
		}

		if time.Since(v.GetStartTime()) < ttl {
			return true
		}

		v.PrepareForDeletion()
		mp.Delete(key)

		return true
	})
}
