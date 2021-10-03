package task

import (
	"log"
	"testing"
	"time"
)

func TestBackgroundRoutine_Timer(t *testing.T) {
	intervalValue := 5 * time.Minute
	//intervalValue := 2 * time.Hour
	offsetValue := time.Duration(0)
	//offsetValue := 11 *time.Hour

	now := time.Now()
	nowOffset := time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second
	nowOffset += offsetValue
	curOffset := (nowOffset/intervalValue+1)*intervalValue - nowOffset

	log.Printf("%v", now)
	log.Printf("%v", curOffset)
}
