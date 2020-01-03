package timer

import (
	"context"
	"sync"
	"sync/atomic"

	unit "github.com/GuoCeng/time-wheel/timer/time-unit"

	"github.com/GuoCeng/time-wheel/queue"
)

type Timer interface {

	/**
	 * Add a new task to this executor. It will be executed after the task's delay
	 * (beginning from the time of submission)
	 * @param timerTask the task to add
	 */
	Add(timerTask Task)

	/**
	   * Advance the internal clock, executing any tasks whose expiration has been
	  *          * reached within the duration of the passed timeout.
	  *          * @param timeoutMs
	  *          * @return whether or not any tasks were executed
	*/
	AdvanceClock(ctx context.Context) bool

	/**
	 * Get the number of tasks pending execution
	 * @return the number of tasks
	 */
	Size() int64

	/**
	 * Shutdown the timer service, leaving pending tasks unexecuted
	 */
	Shutdown()
}

func NewSystemTimer(tickMs int64, wheelSize int) *SystemTimer {
	startMs := unit.ClockMs()
	q := queue.NewDelay()
	c := new(int64)
	return &SystemTimer{
		tickMs:      tickMs,
		wheelSize:   wheelSize,
		startMs:     startMs,
		delayQueue:  q,
		taskCounter: c,
		timingWheel: NewTimingWheel(tickMs, wheelSize, startMs, c, q),
	}
}

type SystemTimer struct {
	mu          sync.RWMutex
	tickMs      int64
	wheelSize   int
	startMs     int64
	delayQueue  *queue.DelayQueue
	taskCounter *int64
	timingWheel *TimingWheel
}

func (t *SystemTimer) Add(task Task) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var entry *TaskEntry
	if entry = task.GetTaskEntry(); entry == nil {
		entry = NewTaskEntry(task, unit.ClockMs()+task.GetDelay())
	} else {
		entry.exp = unit.ClockMs() + task.GetDelay()
	}
	t.addTimerTaskEntry(entry)
}

//将任务插入时间轮中，如果发现任务超时，则执行
func (t *SystemTimer) addTimerTaskEntry(taskEntry *TaskEntry) {
	if !t.timingWheel.add(taskEntry) {
		// Already expired or cancelled
		if !taskEntry.cancelled() {
			go func() {
				taskEntry.task.Run()
			}()
		}
	}
}

// Advances the clock if there is an expired bucket. If there isn't any expired bucket when called,
// waits up to timeoutMs before giving up.
// 收割时间轮，通过延时队列获取对象，如果未返回，则表明未到收割时间，返回的话，就将各圈中的任务进行重新分配，分配过程中将已过期的任务执行掉
func (t *SystemTimer) AdvanceClock(ctx context.Context) bool {
	bucket := t.delayQueue.Pop(ctx)
	if bucket != nil {
		if v, ok := bucket.(*TaskList); ok {
			t.mu.Lock()
			defer t.mu.Unlock()
			for v != nil {
				//推进时间轮时间
				t.timingWheel.advanceClock(unit.HiResClockMs())
				//刷新对象，将时间轮各圈中的对象，重新分配各圈中相应的位置
				entries := v.flush()
				for _, e := range entries {
					if e != nil {
						t.addTimerTaskEntry(e)
					}
				}
				x := t.delayQueue.Poll()
				if x != nil {
					v = x.(*TaskList)
				} else {
					v = nil
				}
			}
			return true
		} else {
			panic("did not found TaskList")
		}
	}
	return false
}

func (t *SystemTimer) Size() int64 {
	return atomic.LoadInt64(t.taskCounter)
}

func (t *SystemTimer) Shutdown() {
	t.delayQueue.Release()
}
