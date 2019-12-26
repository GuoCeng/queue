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
	startMs := unit.HiResClockMs()
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
	t.addTimerTaskEntry(NewTaskEntry(task, task.GetDelay()+unit.HiResClockMs()))
}

func (t *SystemTimer) addTimerTaskEntry(taskEntry *TaskEntry) {
	if !t.timingWheel.add(taskEntry) {
		// Already expired or cancelled
		if !taskEntry.cancelled() {
			//fmt.Printf("任务ID[%v]-当前时间[%v]-任务时间[%v]，已过期，直接执行 \n", taskEntry.task.GetID(), unit.HiResClockMs(), taskEntry.exp)
			go func() {
				taskEntry.task.Run()
			}()
		}
	}
}

// Advances the clock if there is an expired bucket. If there isn't any expired bucket when called,
// waits up to timeoutMs before giving up.
func (t *SystemTimer) AdvanceClock(ctx context.Context) bool {
	bucket := t.delayQueue.Pop(ctx)
	if bucket != nil {
		if v, ok := bucket.(*TaskList); ok {
			t.mu.Lock()
			defer t.mu.Unlock()
			for v != nil {
				t.timingWheel.advanceClock(v.expiration)
				v.flush(func(e *TaskEntry) {
					t.addTimerTaskEntry(e)
				})
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

}
