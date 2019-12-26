package timer

import (
	"sync"

	"github.com/GuoCeng/time-wheel/queue"
)

type TimingWheel struct {
	mu            sync.Mutex
	tickMs        int64
	wheelSize     int
	startMs       int64
	taskCounter   *int64
	q             *queue.DelayQueue
	interval      int64
	buckets       []*TaskList
	currentTime   int64
	overflowWheel *TimingWheel
}

func NewTimingWheel(tickMs int64, wheelSize int, startMs int64, c *int64, q *queue.DelayQueue) *TimingWheel {
	buckets := make([]*TaskList, wheelSize)
	for i := 0; i < wheelSize; i++ {
		buckets[i] = NewTaskList(c)
	}
	timingWheel := &TimingWheel{
		tickMs:      tickMs,
		wheelSize:   wheelSize,
		startMs:     startMs,
		taskCounter: c,
		q:           q,
		interval:    int64(wheelSize) * tickMs,
		buckets:     buckets,
		currentTime: startMs - (startMs % tickMs),
	}
	return timingWheel
}

func (t *TimingWheel) addOverflowWheel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.overflowWheel == nil {
		t.overflowWheel = NewTimingWheel(t.interval, t.wheelSize, t.currentTime, t.taskCounter, t.q)
	}
}

func (t *TimingWheel) add(entry *TaskEntry) bool {
	exp := entry.exp
	if entry.cancelled() {
		// Cancelled
		return false
	} else if exp < t.currentTime+t.tickMs {
		// Already expired
		return false
	} else if exp < t.currentTime+t.interval {
		// Put in its own bucket
		virtualId := exp / t.tickMs
		bucket := t.buckets[virtualId%int64(t.wheelSize)]
		bucket.add(entry)
		if bucket.setExpiration(virtualId * t.tickMs) {
			// The bucket needs to be enqueued because it was an expired bucket
			// We only need to enqueue the bucket when its expiration time has changed, i.e. the wheel has advanced
			// and the previous buckets gets reused; further calls to set the expiration within the same wheel cycle
			// will pass in the same value and hence return false, thus the bucket with the same expiration will not
			// be enqueued multiple times.
			t.q.Offer(bucket)
		}
		return true
	} else {
		// Out of the interval. Put it into the parent timer
		if t.overflowWheel == nil {
			t.addOverflowWheel()
		}
		return t.overflowWheel.add(entry)
	}
}

// Try to advance the clock
func (t *TimingWheel) advanceClock(timeMs int64) {
	if timeMs >= t.currentTime+t.tickMs {
		t.currentTime = timeMs - (timeMs % t.tickMs)
		// Try to advance the clock of the overflow wheel if present
		if t.overflowWheel != nil {
			t.overflowWheel.advanceClock(t.currentTime)
		}
	}
}
