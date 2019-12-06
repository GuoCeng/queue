package queue

import (
	"testing"
	"time"
)

type Demo struct {
	Msg  string
	Time time.Duration
	Exp  time.Time
}

func (d *Demo) GetDelay() time.Time {
	return d.Exp
}

func TestDelayQueue_Pop(t *testing.T) {
	now := time.Now()
	d1 := &Demo{
		Msg:  "demo1",
		Time: 1 * time.Second,
		Exp:  now.Add(1 * time.Second),
	}
	d2 := &Demo{
		Msg:  "demo2",
		Time: 2 * time.Second,
		Exp:  now.Add(2 * time.Second),
	}

	dq := NewDelay()
	go func() {
		g := dq.Pop().(*Demo)
		t.Logf("=============time:%v;g.msg:%v;g.time:%v", time.Now().Second(), g.Msg, g.Time)
	}()
	dq.Offer(d1)
	dq.Offer(d2)
	g2 := dq.Pop().(*Demo)
	t.Logf("time:%v;g.msg:%v;g.time:%v", time.Now().Second(), g2.Msg, g2.Time)

	time.Sleep(1 * time.Second)
}

func TestSelect(t *testing.T) {
	start := time.Now()
	c := make(chan time.Duration)
	select {
	case c <- time.Now().Sub(start):
	default:
		t.Logf("default")
	}
}
