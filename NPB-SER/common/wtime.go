package common

import (
	"time"
)

var start [64]float64
var elapsed [64]float64

func elapsedTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

func TimerClear(n int) {
	elapsed[n] = 0.0
}

func TimerStart(n int) {
	start[n] = elapsedTime()
}

func TimerStop(n int) {
	now := elapsedTime()
	t := now - start[n]
	elapsed[n] += t
}

func TimerRead(n int) float64 {
	return elapsed[n]
}
