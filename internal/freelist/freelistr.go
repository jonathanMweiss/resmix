package freelist

import (
	"bytes"
	"runtime"
	"sync"
)

var (
	ncpus       = runtime.NumCPU() * 3
	regListSize = 256
)

const mb = 1024 * 1024

var List = NewFreelist(regListSize)
var ShortLivedList = NewFreelist(ncpus)

func GetLongLastingBuffer() (*bytes.Buffer, Done) {
	return List.Get()
}

func Get() (*bytes.Buffer, Done) {
	return ShortLivedList.Get()
}

type Freelist struct {
	pool *sync.Pool
}

type Done func()

// Get returns a buffer, and a Done function which can be used to reclaim the buffer.
func (f *Freelist) Get() (*bytes.Buffer, Done) {
	itm := f.pool.Get()
	b := itm.(*bytes.Buffer)
	b.Reset()
	return b, func() { b.Reset(); f.pool.Put(b) }
}

func NewFreelist(numItems int) *Freelist {
	f := &Freelist{pool: &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}}
	for i := 0; i < numItems; i++ {
		f.pool.Put(bytes.NewBuffer(nil))
	}
	return f
}
