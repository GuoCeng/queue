package timer

import (
	"sync"
	"sync/atomic"

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
	id        int
	delayMs   int64
	taskEntry *TaskEntry
	run       func()
}

func (t *SimpleTask) GetID() int64 {
	return t.delayMs
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

func NewSimpleTask(id int, delayMs int64, r func()) *SimpleTask {
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
	Mu   sync.Mutex
	exp  int64
	task Task
	list *TaskList
	next *TaskEntry
	prev *TaskEntry
}

func (t *TaskEntry) clearTask() {
	if t.task != nil {
		t.task.SetTaskEntry(t)
	}
}

func (t *TaskEntry) cancelled() bool {
	return t.task.GetTaskEntry() != t
}

func (t *TaskEntry) Remove() {
	currentList := t.list
	for currentList != nil {
		currentList.remove(t)
		currentList = t.list
	}
}

func NewTaskList(c *int64) *TaskList {
	tl := &TaskList{
		taskCounter: c,
	}
	tl.initRoot()
	return tl
}

type TaskList struct {
	mu          sync.Mutex
	flushMu     sync.Mutex
	removeMu    sync.Mutex
	taskCounter *int64
	root        *TaskEntry
	expiration  int64
}

func (t *TaskList) initRoot() {
	t.root = &TaskEntry{}
	t.root.next = t.root
	t.root.prev = t.root
}

func (t *TaskList) setExpiration(e int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	old := t.expiration
	t.expiration = e
	return old != e
}

func (t *TaskList) Foreach(f func(task Task)) {
	entry := t.root.next
	for entry != t.root {
		nextEntry := entry.next
		if !entry.cancelled() {
			f(entry.task)
		}
		entry = nextEntry
	}
}

func (t *TaskList) add(task *TaskEntry) {
	task.Remove()
	t.mu.Lock()
	defer t.mu.Unlock()
	task.Mu.Lock()
	defer task.Mu.Unlock()
	if task.list == nil {
		tail := t.root.prev
		task.next = t.root
		task.prev = tail
		task.list = t
		tail.next = task
		t.root.prev = task
		atomic.AddInt64(t.taskCounter, 1)
	}
}

func (t *TaskList) remove(task *TaskEntry) {
	t.removeMu.Lock()
	defer t.removeMu.Unlock()
	task.Mu.Lock()
	defer task.Mu.Unlock()
	if task.list == t {
		task.next.prev = task.prev
		task.prev.next = task.next
		task.next = nil
		task.prev = nil
		task.list = nil
		atomic.AddInt64(t.taskCounter, -1)
	}
}

var empty int64 = -1

func (t *TaskList) flush(f func(e *TaskEntry)) {
	t.flushMu.Lock()
	defer t.flushMu.Unlock()
	head := t.root.next
	for head != t.root {
		t.remove(head)
		f(head)
		head = t.root.next
	}
	t.expiration = empty
}

func (t *TaskList) GetDelay(u *unit.TimeUnit) int64 {
	return u.Convert(max(t.expiration-unit.HiResClockMs(), 0), unit.MILLISECONDS)
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
