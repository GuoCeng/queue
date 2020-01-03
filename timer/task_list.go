package timer

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	unit "github.com/GuoCeng/time-wheel/timer/time-unit"
)

type Task interface {
	GetID() int64
	GetDelay() int64
	Cancel()
	SetTaskEntry(entry *TaskEntry)
	GetTaskEntry() *TaskEntry
	Run()
}

type SimpleTask struct {
	mu        sync.RWMutex
	id        int64
	delayMs   int64
	taskEntry *TaskEntry
	run       func()
}

func (t *SimpleTask) GetID() int64 {
	return t.id
}

func (t *SimpleTask) GetDelay() int64 {
	return t.delayMs
}

func (t *SimpleTask) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.taskEntry != nil {
		t.taskEntry.Remove()
	}
	t.taskEntry = nil
}

func (t *SimpleTask) SetTaskEntry(entry *TaskEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.taskEntry != nil && t.taskEntry != entry {
		t.taskEntry.Remove()
	}
	t.taskEntry = entry
}

func (t *SimpleTask) GetTaskEntry() *TaskEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.taskEntry
}

func (t *SimpleTask) Run() {
	t.run()
}

func NewSimpleTask(id int64, delayMs int64, r func()) *SimpleTask {
	return &SimpleTask{
		id:      id,
		delayMs: delayMs,
		run:     r,
	}
}

func NewTaskEntry(task Task, exp int64) *TaskEntry {
	taskEntry := &TaskEntry{
		exp:  exp,
		task: task,
	}
	taskEntry.clearTask()
	return taskEntry
}

type TaskEntry struct {
	mu   sync.Mutex
	exp  int64 //毫秒数
	task Task
	list *TaskList
}

func (t *TaskEntry) clearTask() {
	if t.task != nil {
		t.task.SetTaskEntry(t)
	}
}

func (t *TaskEntry) setList(list *TaskList) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.list = list
}

func (t *TaskEntry) Remove() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.list != nil {
		t.list.remove(t)
		t.list = nil
	}
}

func (t *TaskEntry) cancelled() bool {
	return t.task.GetTaskEntry() != t
}

func NewTaskList(c *int64) *TaskList {
	tl := &TaskList{
		taskCounter: c,
	}
	return tl
}

var mu sync.Mutex

type TaskList struct {
	mu          sync.Mutex
	taskCounter *int64
	entries     []*TaskEntry
	expiration  int64
}

// 如果两个时间放到了时间轮的相同层的相同刻度中，刷新过期时间时，要比较是否比之前的过期时间小，如果小的话才更新，
// 因为如果把大的过期时间更新上去，会导致获取到TaskList时，里面的所有任务都过期了，
// 这样子在N层时间轮之后，时间间隔较大的地方，会出现相差很久的两个任务，却同时执行了
// 返回值为是否已经放入了延时队列中，防止重复放入
func (t *TaskList) setExpiration(e int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	old := t.expiration
	if old == 0 || old == empty || (old != empty && old > e) {
		t.expiration = e
	}
	return old == e
}

func (t *TaskList) add(entry *TaskEntry) {
	mu.Lock()
	defer mu.Unlock()
	entry.setList(t)
	atomic.AddInt64(t.taskCounter, 1)
	t.entries = append(t.entries, entry)
}

func (t *TaskList) remove(entry *TaskEntry) {
	mu.Lock()
	defer mu.Unlock()
	if entry.list == t {
		for idx, e := range t.entries {
			if e == entry {
				t.entries[idx] = nil
				e.list = nil
				t.entries = append(t.entries[:idx], t.entries[idx+1:]...)
				break
			}
		}
		atomic.AddInt64(t.taskCounter, -1)
	}
}

var empty int64 = math.MaxInt64

func (t *TaskList) flush() []*TaskEntry {
	mu.Lock()
	var entries = make([]*TaskEntry, len(t.entries))
	copy(entries, t.entries)
	mu.Unlock()
	for _, entry := range entries {
		t.remove(entry)
	}
	t.expiration = empty
	return entries
}

func (t *TaskList) GetDelay() time.Duration {
	return time.Duration(t.expiration-unit.HiResClockMs()) * time.Millisecond
}
