package queue

import (
	"context"
	"testing"
	"time"
)

type Demo struct {
	Msg  string
	Time time.Duration
	Exp  time.Time
}

func (d *Demo) GetDelay() time.Duration {
	now := time.Now()
	if now.After(d.Exp) {
		return -1
	} else {
		return d.Exp.Sub(now)
	}
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
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	go func() {
		g := dq.Pop(ctx).(*Demo)
		t.Logf("=============time:%v;g.msg:%v;g.time:%v", time.Now().Second(), g.Msg, g.Time)
	}()
	dq.Offer(d1)
	dq.Offer(d2)
	g2 := dq.Pop(ctx).(*Demo)
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
