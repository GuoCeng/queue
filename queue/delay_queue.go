package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Delayed interface {
	GetDelay() time.Duration
}

func NewDelay() *DelayQueue {

	return &DelayQueue{
		available: make(chan struct{}, 1),
		count:     new(int64),
		q:         NewPriority(),
	}
}

type DelayQueue struct {
	mu        sync.Mutex
	count     *int64
	available chan struct{}
	q         *PriorityQueue
}

func (dq *DelayQueue) Offer(e Delayed) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	i := &Item{
		Value:    e,
		Priority: func() int64 { return int64(e.GetDelay()) },
	}
	dq.q.Push(i)
	count := atomic.AddInt64(dq.count, 1)
	if dq.q.Len() == 1 && count == 1 {
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
			delay := delayed.GetDelay()
			delayTime = delay
			//精度设置为毫秒，与时间轮的精度毫秒匹配，不然容易出现延时任务执行时间出现偏差
			if delay < 1*time.Millisecond {
				defer dq.mu.Unlock()
				item := dq.q.Pop().(*Item)
				atomic.AddInt64(dq.count, -1)
				return item.Value
			} else {
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
			delay := delayed.GetDelay()
			if delay <= 0 {
				item := dq.q.Pop().(*Item)
				atomic.AddInt64(dq.count, -1)
				return item.Value
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
}

func (dq *DelayQueue) Release() {

}
