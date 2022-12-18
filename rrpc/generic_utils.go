package rrpc

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type withAnswer interface {
	GetStartTime() time.Time
	PublishResponse(error)
}

func (r *requestWithResponse[T]) GetStartTime() time.Time {
	return r.StartTime
}

func (r *requestWithResponse[T]) PublishResponse(err error) {
	r.Err <- err
	close(r.Err)
}

func garbageCollectSyncMap[S withAnswer](ctx context.Context, wg *sync.WaitGroup, mp *sync.Map) {
	defer wg.Done()
	dropFromMap := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-dropFromMap.C:
			currentTime := time.Now()
			mp.Range(func(key, value interface{}) bool {
				rqst, ok := value.(S) // .Err <- fmt.Errorf("timeout"))
				if !ok {
					panic("client::VerifyAndDispatch: could not cast task!")
				}
				if currentTime.Sub(rqst.GetStartTime()) > time.Second*5 {
					rqst.PublishResponse(status.Error(codes.Canceled, "response timed out"))
					mp.Delete(key)
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}
