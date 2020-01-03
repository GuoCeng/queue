package timer

import (
	"sync"

	unit "github.com/GuoCeng/time-wheel/timer/time-unit"

	"github.com/GuoCeng/time-wheel/queue"
)

type TimingWheel struct {
	mu            sync.Mutex
	tickMs        int64             //刻度（精度毫秒）
	wheelSize     int               //时间轮每圈的大小
	startMs       int64             //开始时间（单位毫秒）
	taskCounter   *int64            //总任务数
	q             *queue.DelayQueue //延时队列
	interval      int64             //当前圈的时间跨度（单位毫秒）
	buckets       []*TaskList       //当前圈的任务列表
	currentTime   int64             //当前圈保持的当前时间（由时间轮进行推进）
	overflowWheel *TimingWheel      //超过当前圈时间跨度时，会创建新的圈
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
		interval:    tickMs * int64(wheelSize),
		buckets:     buckets,
		currentTime: startMs - (startMs % tickMs),
	}
	return timingWheel
}

//
func (t *TimingWheel) addOverflowWheel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.overflowWheel == nil {
		t.overflowWheel = NewTimingWheel(t.interval, t.wheelSize, t.currentTime, t.taskCounter, t.q)
	}
}

//添加任务，将任务添加到对应时间轮的圈的位置
func (t *TimingWheel) add(entry *TaskEntry) bool {
	expiration := entry.exp
	if entry.cancelled() {
		// Cancelled
		return false
	} else if expiration < t.currentTime+t.tickMs {
		// Already expired
		return false
	} else if expiration < t.currentTime+t.interval {
		// Put in its own bucket
		virtualId := (expiration - t.currentTime) / t.tickMs
		bucket := t.buckets[int64(virtualId)%int64(t.wheelSize)]
		bucket.add(entry)
		// 设置延时队列对象的超时时间，用当前添加对象的时间作为超时时间
		if !bucket.setExpiration(unit.HiResClockMs() + virtualId*t.tickMs) {
			// The bucket needs to be enqueued because it was an expired bucket
			// We only need to enqueue the bucket when its expiration time has changed, i.e. the wheel has advanced
			// and the previous buckets gets reused; further calls to set the expiration within the same wheel cycle
			// will pass in the same value and hence return false, thus the bucket with the same expiration will not
			// be enqueued multiple times.
			// 插入延时队列
			t.q.Offer(bucket)
		}
		return true
	} else {
		// Out of the interval. Put it into the parent timer
		// 如果超过当前圈时间跨度，则将该任务插入下一圈中
		if t.overflowWheel == nil {
			t.addOverflowWheel()
		}
		return t.overflowWheel.add(entry)
	}
}

// Try to advance the clock
//推进时间轮的当前时间
func (t *TimingWheel) advanceClock(timeMs int64) {
	if timeMs >= t.currentTime+t.tickMs {
		t.currentTime = timeMs - (timeMs % t.tickMs)
		// Try to advance the clock of the overflow wheel if present
		if t.overflowWheel != nil {
			t.overflowWheel.advanceClock(t.currentTime)
		}
	}
}
