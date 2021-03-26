package twitchchat

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

var ErrBucketClosed = fmt.Errorf("Error: Sink Closed")

// inspiration from github.com/Docker/go-events
type Event interface{}

// Emitter accepts and emits events
type Emitter interface {
	// Emit event
	Emit(event Event) error

	OnError(err error)

	Close() error
}

// Bucket controls the flow of events into the sink
type Bucket struct {
	emitter Emitter
	events  *list.List
	cond    *sync.Cond
	mutex   sync.Mutex
	limiter *rate.Limiter
	context context.Context
	closed  bool
}

// Makes a new bucket that can be filled with events. Events are dripped at the
// passed in rate with given burstLimit. To have no rate limit, rate.Inf should be
// passed in
func NewBucket(emitter Emitter, tokenRate rate.Limit, burstLimit int) *Bucket {
	bucket := Bucket{
		emitter: emitter,
		events:  list.New(),
		limiter: rate.NewLimiter(tokenRate, burstLimit),
		context: context.Background(),
	}
	bucket.cond = sync.NewCond(&bucket.mutex)

	go bucket.drip()
	return &bucket
}

// Add event into bucket. If it's high priority the event will
// be pushed to the front of the list
func (bucket *Bucket) AddEvent(event Event, highPriority bool) error {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	if bucket.closed {
		return ErrBucketClosed
	}

	if highPriority {
		bucket.events.PushFront(event)
	} else {
		bucket.events.PushBack(event)
	}
	bucket.cond.Signal()

	return nil
}

func (bucket *Bucket) Close() error {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	if bucket.closed {
		return nil
	}

	bucket.closed = true
	bucket.cond.Signal()
	bucket.cond.Wait()
	return bucket.emitter.Close()
}

func (bucket *Bucket) drip() {
	for {
		event := bucket.next()
		if event == nil {
			return
		}

		// Wait for a token before emitting event
		err := bucket.limiter.Wait(bucket.context)
		if err != nil {
			return
		}

		if err := bucket.emitter.Emit(event); err != nil {
			bucket.emitter.OnError(err)
		}
	}
}

func (bucket *Bucket) next() Event {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	for bucket.events.Len() < 1 {
		bucket.cond.Wait()
	}

	front := bucket.events.Front()
	event := front.Value.(Event)
	bucket.events.Remove(front)

	return event
}
