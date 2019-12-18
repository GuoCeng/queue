package timer

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSystemTimer_AdvanceClock(t *testing.T) {
	timer := NewSystemTimer(time.Microsecond, 1000)
	fmt.Printf("开始时间：%v \n", time.Now())
	var delay time.Duration
	/*for i := 0; i < 10; i++ {
		x := i + 1
		r := rand.Intn(100)
		var t time.Duration
		if r == 0 {
			t = time.Duration(x) * time.Second
		} else {
			t = time.Duration(r) * time.Duration(1000) * time.Duration(x) * time.Microsecond
		}
		delay += t
		fmt.Printf("%v-延时时间：%v \n", x, t)
		task := NewTask(t, func() {
			fmt.Printf("%v-执行时间：%v \n", x, time.Now())
		})
		timer.Add(task)
		if t.Seconds() > 5 {
			fmt.Printf("%v-延时时间：%v，超过5S，取消任务 \n", x, t)
			task.Cancel()
		}
	}*/

	timer.Add(NewTask(1*time.Second, func() {
		fmt.Printf("%v-执行时间：%v \n", 11, time.Now())
	}))

	timer.Add(NewTask(3*24*time.Hour, func() {
		fmt.Printf("%v-执行时间：%v \n", 11, time.Now())
	}))

	delay += 3 * 24 * time.Hour

	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("====================")
		timer.Add(NewTask(1*time.Second, func() {
			fmt.Printf("%v-执行时间：%v \n", 22, time.Now())
		}))
	}()
	go func() {
		for {
			ctx, _ := context.WithTimeout(context.Background(), 1/5*time.Microsecond)
			timer.AdvanceClock(ctx)
		}
	}()
	time.Sleep(24 * time.Hour)
}
