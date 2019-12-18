package timer

import (
	"sync"
	"time"
)

type Task struct {
	mu        sync.Mutex
	delayMs   time.Duration
	taskEntry *TaskEntry
	run       func()
}

func (t *Task) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.taskEntry != nil {
		t.taskEntry.remove()
	}
	t.taskEntry = nil
}

func (t *Task) setTaskEntry(entry *TaskEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.taskEntry != nil && t.taskEntry != entry {
		t.taskEntry.remove()
	}
	t.taskEntry = entry
}

func NewTask(delayMs time.Duration, r func()) *Task {
	return &Task{
		delayMs: delayMs,
		run:     r,
	}
}

func NewTaskEntry(task *Task, exp time.Time) *TaskEntry {
	taskEntry := &TaskEntry{
		exp:  exp,
		task: task,
	}
	taskEntry.clearTask()
	return taskEntry
}

type TaskEntry struct {
	exp  time.Time
	task *Task
	list *TaskList
	next *TaskEntry
	prev *TaskEntry
}

func (t *TaskEntry) clearTask() {
	if t.task != nil {
		t.task.setTaskEntry(t)
	}
}

func (t *TaskEntry) cancelled() bool {
	return t.task.taskEntry != t
}

func (t *TaskEntry) remove() {
	currentList := t.list
	for currentList != nil {
		currentList.remove(t)
		currentList = t.list
	}
}

/*func (t *TaskEntry) compare(x *TaskEntry) int {
	if t.expirationMs > x.expirationMs {
		return 1
	} else if t.expirationMs == x.expirationMs {
		return 0
	} else {
		return -1
	}

}*/

func NewTaskList(taskCounter int64) *TaskList {
	tl := &TaskList{
		taskCounter: taskCounter,
	}
	tl.initRoot()
	return tl
}

type TaskList struct {
	mu          sync.Mutex
	flushMu     sync.Mutex
	removeMu    sync.Mutex
	taskCounter int64
	root        *TaskEntry
	expiration  time.Time
}

func (t *TaskList) initRoot() {
	t.root = &TaskEntry{}
	t.root.next = t.root
	t.root.prev = t.root
}

func (t *TaskList) setExpiration(e time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	old := t.expiration
	t.expiration = e
	return old != e
}

func (t *TaskList) Foreach(f func(task *Task)) {
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
	t.mu.Lock()
	defer t.mu.Unlock()
	task.remove()
	if task.list == nil {
		tail := t.root.prev
		task.next = t.root
		task.prev = tail
		task.list = t
		tail.next = task
		t.root.prev = task
		t.taskCounter++
	}
}

func (t *TaskList) remove(task *TaskEntry) {
	t.removeMu.Lock()
	defer t.removeMu.Unlock()
	if task.list == t {
		task.next.prev = task.prev
		task.prev.next = task.next
		task.next = nil
		task.prev = nil
		task.list = nil
		t.taskCounter--
	}
}

var empty = time.Time{}

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

func (t *TaskList) GetDelay() time.Duration {
	now := time.Now()
	if now.After(t.expiration) {
		return -1
	}
	return t.expiration.Sub(now)
}

/*func (t *TaskList) compareTo(d queue.Delayed) int {
	tl, ok := d.(*TaskList)
	if !ok {
		panic("can not convert to TaskList")
	}
	if t.expiration.Before(tl.expiration) {
		return -1
	} else if t.expiration.After(tl.expiration) {
		return 1
	} else {
		return 0
	}
}*/
