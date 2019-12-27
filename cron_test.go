package time_wheel

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	"github.com/GuoCeng/time-wheel/cron"
)

func TestCron(t *testing.T) {
	go func() {
		ip := "0.0.0.0:6060"
		fmt.Printf("start pprof success on %s\n", ip)
		if err := http.ListenAndServe(ip, nil); err != nil {
			fmt.Printf("start pprof failed on %s\n", ip)
		}
	}()
	cron := cron.New()
	cron.AddFunc("3 * * * * *", func() {
		log.Println("3-time")
	})
	cron.AddFunc("30 * * * * *", func() {
		log.Println("30-time")
	})
	cron.AddFunc("0/5 * * * * *", func() {
		log.Println("0/5-time")
	})

	cron.AddFunc("1/10 * * * * *", func() {
		log.Println("1/10-time")
	})
	fmt.Printf("开始时间：%v \n", time.Now())
	cron.Run()
}
