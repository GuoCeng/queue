package timer

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestSystemTimer_AdvanceClock(t *testing.T) {
	timer := NewSystemTimer(time.Microsecond, 1000)
	fmt.Printf("开始时间：%v \n", time.Now())
	var delay time.Duration
	for i := 0; i < 10; i++ {
		x := i + 1
		r := rand.Intn(10)
		var t time.Duration
		if r == 0 {
			t = time.Duration(x) * time.Second
		} else {
			t = time.Duration(r) * time.Duration(1000) * time.Duration(x) * time.Microsecond
		}
		delay += t
		fmt.Printf("%v-延时时间：%vs \n", x, t)
		timer.Add(NewTask(t, func() {
			fmt.Printf("%v-执行时间：%v \n", x, time.Now())
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(delay)
		cancel()
		fmt.Printf("完成时间：%v \n", time.Now())
	}()
	timer.AdvanceClock(ctx)
}
