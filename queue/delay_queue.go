package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Delayed interface {
	GetDelay() time.Duration
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
	delay := e.GetDelay()
	i := &Item{
		Value:    e,
		Priority: int64(delay),
	}
	dq.q.Push(i)
	if dq.q.Len() == 1 {
		go func() {
			dq.available <- struct{}{}
			fmt.Println("dq.available doing")
		}()
	}
}

func (dq *DelayQueue) Pop(ctx context.Context) interface{} {
	var delayTime time.Duration
Start:
	dq.mu.Lock()
	first, ok := dq.q.Peek().(*Item)
	if !ok {
		dq.mu.Unlock()
		goto Wait
	} else {
		if delayed, ok := first.Value.(Delayed); ok {
			delay := delayed.GetDelay()
			if delay <= 0 {
				dq.mu.Unlock()
				item := dq.q.Pop().(*Item)
				return item.Value
			} else {
				delayTime = delay
				dq.mu.Unlock()
				goto Wait
			}
		}
	}

Wait:
	select {
	case <-ctx.Done():
		return nil
	case <-dq.available:
		fmt.Println("get dq.available")
		goto Start
	case <-time.After(delayTime):
		goto Start
	}
}
