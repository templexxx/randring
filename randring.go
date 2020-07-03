// Package randring provides a ring buckets for multi-producer & one-consumer
// which will drop messages if buckets full and won't guarantee order.
// randring only cares about memory corruption.
package randring

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/templexxx/cpu"
)

const falseSharingRange = cpu.X86FalseSharingRange

type bucket struct {
	//_padding0 [falseSharingRange]byte
	data unsafe.Pointer
	//_padding1 [falseSharingRange]byte
}

// Ring provides a ring buckets for multi-producer & one-consumer.
type Ring struct {
	mask       uint64
	_          [falseSharingRange]byte
	writeIndex uint64
	_          [falseSharingRange]byte

	// writeIndex cache for Pop, only get new write index when read catch write.
	// Help to reduce caching missing.
	writeIndexCache uint64
	_               [falseSharingRange]byte
	readIndex       uint64

	buckets []bucket
}

// New creates a ring.
// ring size = 2 ^ n.
func New(n uint64) *Ring {

	if n > 16 || n == 0 {
		panic("illegal ring size")
	}

	r := &Ring{
		buckets: make([]bucket, 1<<n),
		mask:    (1 << n) - 1,
	}

	r.writeIndex = ^r.writeIndex
	return r
}

// Push puts the data in ring in the next bucket no matter what in it.
func (r *Ring) Push(data unsafe.Pointer) {
	idx := atomic.AddUint64(&r.writeIndex, 1) & r.mask
	atomic.StorePointer(&r.buckets[idx].data, data)
	time.Now().UnixNano()
}

// TryPop tries to pop data from the next bucket,
// return (nil, false) if no data available.
func (r *Ring) TryPop() (unsafe.Pointer, bool) {

	if r.readIndex > r.writeIndexCache {
		r.writeIndexCache = atomic.LoadUint64(&r.writeIndex)
		if r.readIndex > r.writeIndexCache {
			return nil, false
		}
	}

	idx := r.readIndex & r.mask
	data := atomic.SwapPointer(&r.buckets[idx].data, nil)

	if data == nil {
		return nil, false
	}

	r.readIndex++
	return data, true
}
