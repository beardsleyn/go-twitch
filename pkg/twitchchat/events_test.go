package twitchchat

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

type testEmitter struct {
	Emitter
	events []Event
}

func newTestEmitter(maxEvents int) *testEmitter {
	return &testEmitter{
		events: make([]Event, 0, maxEvents),
	}
}

func (te *testEmitter) Emit(event Event) error {
	te.events = append(te.events, event)
	return nil
}

func (te *testEmitter) OnError(err error) {
	// nothing to do
}

func (te *testEmitter) Close() error {
	// nothing to do
	return nil
}

func TestBucket(t *testing.T) {
	numEvents := 10
	rateTime := 1000 * time.Millisecond
	rate := rate.Every(rateTime)
	burst := 2

	emitter := newTestEmitter(numEvents)
	bucket := NewBucket(emitter, rate, burst)

	// Let goroutines settle
	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(event Event) {
			if err := bucket.AddEvent(event, false); err != nil {
				t.Fatalf("Error writing event")
			}
			wg.Done()
		}("event" + fmt.Sprint((i)))
	}
	wg.Wait()

	for i := burst; i < numEvents/burst-1; i = i + burst {
		if len(emitter.events) != i {
			t.Fatalf("Num events: " + fmt.Sprint(len(emitter.events)) + " expected: " + fmt.Sprint(i))
		}

		time.Sleep(rateTime)
	}

}
