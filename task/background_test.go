package task

import (
	"log"
	"testing"
	"time"
)

func TestBackgroundRoutine_Timer(t *testing.T) {
	intervalValue := 5 * time.Minute
	offsetValue := time.Duration(0)

	now := time.Now()
	nowOffset := time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second
	nowOffset += offsetValue
	curOffset := (nowOffset/intervalValue+1)*intervalValue - nowOffset - offsetValue

	log.Printf("%v", now)
	log.Printf("%v", curOffset)
}
