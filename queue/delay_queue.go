package queue

import (
	"sync"
	"time"
)

type Delayed interface {
	GetDelay() time.Duration
}

type DelayQueue struct {
	mu        sync.Mutex
	available chan struct{}
	q         *PriorityQueue
}

func (dq DelayQueue) Offer(e Delayed) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	i := &Item{
		Value:    e,
		Priority: int64(e.GetDelay()),
	}
	dq.q.Push(i)
	if dq.q.Len() == 1 {
		dq.available <- struct{}{}
	}
}

func (dq DelayQueue) Pop() interface{} {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	for {
		first, ok := dq.q.Peek().(*Item)
		if !ok {
			select {
			case <-dq.available:
				continue
			}
		} else {
			if delayed, ok := first.Value.(Delayed); ok {
				d := delayed.GetDelay()
				if d.Milliseconds() < 0 {
					return dq.q.Pop()
				} else {

				}
			}
		}
	}
}
