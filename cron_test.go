package time_wheel

import (
	"fmt"
	"testing"
	"time"

	"github.com/GuoCeng/time-wheel/cron"
)

func TestCron(t *testing.T) {
	cron := cron.New()
	cron.AddFunc("1 * * * * *", func() {
		fmt.Printf("1-time: %v \n", time.Now())
	})
	cron.AddFunc("7/30 * * * * *", func() {
		fmt.Printf("7/30-time: %v \n", time.Now())
	})
	cron.AddFunc("0/5 * * * * *", func() {
		fmt.Printf("0/5-time: %v \n", time.Now())
	})

	cron.AddFunc("6 * * * * *", func() {
		fmt.Printf("6-time: %v \n", time.Now())
	})
	fmt.Printf("开始时间：%v \n", time.Now())
	cron.Run()
}
