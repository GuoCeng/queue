package timer

import (
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	mu sync.Mutex
	delayMs time.Duration
	taskEntry *TaskEntry
}

func (t *Task) cancel() {
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

type TaskEntry struct {
	expirationMs time.Duration
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
	for ;currentList != nil; {
		//currentList.remove(this)
		currentList = t.list
	}
}

func (t *TaskEntry) compare(x *TaskEntry) int {
	if t.expirationMs > x.expirationMs {
		return 1
	} else if t.expirationMs == x.expirationMs {
		return 0
	} else {
		return -1
	}

}

type TaskList struct {
	taskCounter int64
	root *TaskEntry
	expiration time.Duration
}

func (t *TaskList) initRoot() {
	t.root = &TaskEntry{}
	t.root.next = t.root
	t.root.prev = t.root
}

func (t *TaskList) setExpiration(e time.Duration) bool {
	atomic.t.expiration.(e)
}
private[timer] class TimerTaskList(taskCounter: AtomicInteger) extends Delayed {

// TimerTaskList forms a doubly linked cyclic list using a dummy root entry
// root.next points to the head
// root.prev points to the tail
private[this] val root = new TimerTaskEntry(null, -1)
root.next = root
root.prev = root

private[this] val expiration = new AtomicLong(-1L)

// Set the bucket's expiration time
// Returns true if the expiration time is changed
def setExpiration(expirationMs: Long): Boolean = {
expiration.getAndSet(expirationMs) != expirationMs
}

// Get the bucket's expiration time
def getExpiration(): Long = {
expiration.get()
}

// Apply the supplied function to each of tasks in this list
def foreach(f: (TimerTask)=>Unit): Unit = {
synchronized {
var entry = root.next
while (entry ne root) {
val nextEntry = entry.next

if (!entry.cancelled) f(entry.timerTask)

entry = nextEntry
}
}
}

// Add a timer task entry to this list
def add(timerTaskEntry: TimerTaskEntry): Unit = {
var done = false
while (!done) {
// Remove the timer task entry if it is already in any other list
// We do this outside of the sync block below to avoid deadlocking.
// We may retry until timerTaskEntry.list becomes null.
timerTaskEntry.remove()

synchronized {
timerTaskEntry.synchronized {
if (timerTaskEntry.list == null) {
// put the timer task entry to the end of the list. (root.prev points to the tail entry)
val tail = root.prev
timerTaskEntry.next = root
timerTaskEntry.prev = tail
timerTaskEntry.list = this
tail.next = timerTaskEntry
root.prev = timerTaskEntry
taskCounter.incrementAndGet()
done = true
}
}
}
}
}

// Remove the specified timer task entry from this list
def remove(timerTaskEntry: TimerTaskEntry): Unit = {
synchronized {
timerTaskEntry.synchronized {
if (timerTaskEntry.list eq this) {
timerTaskEntry.next.prev = timerTaskEntry.prev
timerTaskEntry.prev.next = timerTaskEntry.next
timerTaskEntry.next = null
timerTaskEntry.prev = null
timerTaskEntry.list = null
taskCounter.decrementAndGet()
}
}
}
}

// Remove all task entries and apply the supplied function to each of them
def flush(f: (TimerTaskEntry)=>Unit): Unit = {
synchronized {
var head = root.next
while (head ne root) {
remove(head)
f(head)
head = root.next
}
expiration.set(-1L)
}
}

def getDelay(unit: TimeUnit): Long = {
unit.convert(max(getExpiration - Time.SYSTEM.hiResClockMs, 0), TimeUnit.MILLISECONDS)
}

def compareTo(d: Delayed): Int = {

val other = d.asInstanceOf[TimerTaskList]

if(getExpiration < other.getExpiration) -1
else if(getExpiration > other.getExpiration) 1
else 0
}

}


