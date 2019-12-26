package timer

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestSystemTimer_AdvanceClock(t *testing.T) {
	timer := NewSystemTimer(1, 20)
	fmt.Printf("开始时间：%v \n", time.Now())

	timer.Add(NewSimpleTask(1, 2000, func() {
		fmt.Printf("%v-执行时间：%v \n", 1, time.Now())
	}))

	timer.Add(NewSimpleTask(2, 3000, func() {
		fmt.Printf("%v-执行时间：%v \n", 2, time.Now())
	}))

	timer.Add(NewSimpleTask(3, 4000, func() {
		fmt.Printf("%v-执行时间：%v \n", 3, time.Now())
	}))

	timer.Add(NewSimpleTask(4, 5000, func() {
		fmt.Printf("%v-执行时间：%v \n", 4, time.Now())
	}))

	timer.Add(NewSimpleTask(5, 6000, func() {
		fmt.Printf("%v-执行时间：%v \n", 5, time.Now())
	}))

	go func() {
		fmt.Printf("task counter: %v\n", timer.Size())
		for timer.Size() > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			timer.AdvanceClock(ctx)
			log.Println("执行一次收割", timer.timingWheel.currentTime)
			cancel()
		}
		fmt.Printf("task counter: %v\n", timer.Size())
	}()
	time.Sleep(60 * time.Second)
}
