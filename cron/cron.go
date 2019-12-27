package cron

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuoCeng/time-wheel/timer"
)

type Cron struct {
	entries   map[EntryID]*Entry
	chain     Chain
	stop      chan struct{}
	cycle     chan *Entry
	running   bool
	runningMu sync.Mutex
	parser    ScheduleParser
	nextID    *EntryID
	timer     timer.Timer
}

// Schedule describes a job's duty cycle.
type Schedule interface {
	// Next returns the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

type ScheduleParser interface {
	Parse(spec string) (Schedule, error)
}

type Job interface {
	Run()
}

type EntryID = int64

type Entry struct {
	tmu        sync.RWMutex
	emu        sync.RWMutex
	ID         EntryID
	Schedule   Schedule
	delayMs    int64
	Next       time.Time
	Prev       time.Time
	WrappedJob Job
	Job        Job
	taskEntry  *timer.TaskEntry
	cron       *Cron
}

// Valid returns true if this is not the zero entry.
func (e *Entry) Valid() bool { return e.GetID() != 0 }

func (t *Entry) GetID() int64 {
	return t.ID
}

func (e *Entry) GetDelay() int64 {
	e.tmu.RLock()
	defer e.tmu.RUnlock()
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local)
	next := e.Schedule.Next(t)
	e.Prev = e.Next
	e.Next = next
	e.delayMs = next.Sub(t).Milliseconds()
	return e.delayMs
}
func (e *Entry) Cancel() {
	e.emu.Lock()
	defer e.emu.Unlock()
	if e.taskEntry != nil {
		e.taskEntry.Remove()
	}
	e.taskEntry = nil
}

func (e *Entry) SetTaskEntry(entry *timer.TaskEntry) {
	e.emu.Lock()
	defer e.emu.Unlock()
	if e.taskEntry != nil && e.taskEntry != entry {
		e.taskEntry.Remove()
	}
	e.taskEntry = entry
}
func (e *Entry) GetTaskEntry() *timer.TaskEntry {
	e.emu.RLock()
	defer e.emu.RUnlock()
	return e.taskEntry
}

func (e *Entry) Run() {
	e.WrappedJob.Run()
	e.cron.timer.Add(e)
	//log.Println("run", "now", time.Now(), "entry", e.ID, "next", e.Next)
}

func New(opts ...Option) *Cron {
	c := &Cron{
		entries:   make(map[EntryID]*Entry),
		chain:     NewChain(),
		stop:      make(chan struct{}),
		running:   false,
		cycle:     make(chan *Entry),
		runningMu: sync.Mutex{},
		parser:    standardParser,
		nextID:    new(EntryID),
		timer:     timer.NewSystemTimer(1000, 60),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FuncJob is a wrapper that turns a func() into a cron.Job
type FuncJob func()

func (f FuncJob) Run() { f() }

// AddFunc adds a func to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque GetID is returned that can be used to later remove it.
func (c *Cron) AddFunc(spec string, cmd func()) (EntryID, error) {
	return c.AddJob(spec, FuncJob(cmd))
}

// AddJob adds a Job to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque GetID is returned that can be used to later remove it.
func (c *Cron) AddJob(spec string, cmd Job) (EntryID, error) {
	schedule, err := c.parser.Parse(spec)
	if err != nil {
		return 0, err
	}
	return c.Schedule(schedule, cmd), nil
}

// Schedule adds a Job to the Cron to be run on the given schedule.
// The job is wrapped with the configured Chain.
func (c *Cron) Schedule(schedule Schedule, cmd Job) EntryID {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	nextID := atomic.AddInt64(c.nextID, 1)
	entry := &Entry{
		ID:         nextID,
		Schedule:   schedule,
		WrappedJob: c.chain.Then(cmd),
		Job:        cmd,
		cron:       c,
	}
	c.entries[nextID] = entry
	c.timer.Add(entry)
	return entry.ID
}

// Entry returns a snapshot of the given entry, or nil if it couldn't be found.
func (c *Cron) Entry(id EntryID) *Entry {
	if e, ok := c.entries[id]; ok {
		return e
	}
	return &Entry{}
}

// Remove an entry from being run in the future.
func (c *Cron) Remove(id EntryID) {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	c.Entry(id).Cancel()
	delete(c.entries, id)
}

// Start the cron scheduler in its own goroutine, or no-op if already started.
func (c *Cron) Start() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		return
	}
	c.running = true
	go c.run()
}

// Run the cron scheduler, or no-op if already running.
func (c *Cron) Run() {
	c.runningMu.Lock()
	if c.running {
		c.runningMu.Unlock()
		return
	}
	c.running = true
	c.runningMu.Unlock()
	c.run()
}

// run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Cron) run() {
	for c.running {
		ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
		c.timer.AdvanceClock(ctx)
	}
}

// now returns current time in c location
func (c *Cron) now() time.Time {
	return time.Now()
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
// A context is returned so the caller can wait for running jobs to complete.
func (c *Cron) Stop() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		c.running = false
	}
}
