package queue

import (
	"context"
	"sync"
	"time"

	unit "github.com/GuoCeng/time-wheel/timer/time-unit"
)

type Delayed interface {
	GetDelay(*unit.TimeUnit) int64
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
	delay := e.GetDelay(unit.NANOSECONDS)
	i := &Item{
		Value:    e,
		Priority: delay,
	}
	dq.q.Push(i)
	if dq.q.Len() == 1 {
		go func() {
			dq.available <- struct{}{}
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
			delay := delayed.GetDelay(unit.NANOSECONDS)
			if delay <= 0 {
				dq.mu.Unlock()
				item := dq.q.Pop().(*Item)
				return item.Value
			} else {
				delayTime = time.Duration(delay * unit.MilliScale)
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
		goto Start
	case <-time.After(delayTime):
		goto Start
	}
}

func (dq *DelayQueue) Poll() interface{} {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	first, ok := dq.q.Peek().(*Item)
	if !ok {
		return nil
	} else {
		if delayed, ok := first.Value.(Delayed); ok {
			delay := delayed.GetDelay(unit.NANOSECONDS)
			if delay <= 0 {
				item := dq.q.Pop().(*Item)
				return item.Value
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
}
