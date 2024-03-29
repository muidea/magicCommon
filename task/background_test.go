package task

import (
	"log"
	"testing"
	"time"
)

func TestBackgroundRoutine_Timer(t *testing.T) {
	//intervalValue := 10 * time.Minute
	intervalValue := 24 * time.Hour
	//offsetValue := time.Duration(0)
	offsetValue := 1 * time.Hour

	curOffset := func() time.Duration {
		now := time.Now()
		log.Printf("%v", now)
		nowOffset := time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second
		if intervalValue < 24*time.Hour {
			return (nowOffset/intervalValue+1)*intervalValue - nowOffset
		}

		return (offsetValue + intervalValue - nowOffset + 24*time.Hour) % (24 * time.Hour)
	}()

	log.Printf("%v", curOffset)
}
