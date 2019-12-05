package queue

import (
	"sync"
	"time"
)

type Delayed interface {
	GetDelay() time.Time
}

func NewDelay() *DelayQueue {
	return &DelayQueue{
		available: make(chan struct{}),
		q:         NewPriority(),
	}
}

type DelayQueue struct {
	mu        sync.Mutex
	available chan struct{}
	q         *PriorityQueue
}

func (dq *DelayQueue) Offer(e Delayed) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	var delay time.Duration
	now := time.Now()
	delayed := e.GetDelay()
	if now.Before(delayed) {
		delay = delayed.Sub(now)
	} else {
		delay = 0
	}
	i := &Item{
		Value:    e,
		Priority: int64(delay),
	}
	dq.q.Push(i)
	if dq.q.Len() == 1 {
		go func() {
			dq.available <- struct{}{}
		}()
	}
}

func (dq *DelayQueue) Pop() interface{} {
	var delayTime time.Duration
Start:
	dq.mu.Lock()
	first, ok := dq.q.Peek().(*Item)
	if !ok {
		dq.mu.Unlock()
		goto Wait
	} else {
		if delayed, ok := first.Value.(Delayed); ok {
			now := time.Now()
			delay := delayed.GetDelay()
			if now.After(delay) {
				dq.mu.Unlock()
				item := dq.q.Pop().(*Item)
				return item.Value
			} else {
				delayTime = delay.Sub(now)
				dq.mu.Unlock()
				goto Sleep
			}
		}
	}

Wait:
	<-dq.available
	goto Start

Sleep:
	<-time.After(delayTime)
	goto Start
}
