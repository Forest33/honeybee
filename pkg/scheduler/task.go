package scheduler

import (
	"math"
	"time"
)

const (
	defaultTaskDelay             = time.Second
	defaultTaskMaxDelay          = time.Hour
	defaultTaskMaxProcessingTime = time.Hour * 24 * 7
	defaultExpFactor             = 1
)

type TaskHandler func() error

type Task struct {
	Sender            string
	Delay             time.Duration
	MaxAttempts       int
	MaxProcessingTime time.Duration
	ExpFactor         float64
	Handler           TaskHandler
	id                int64
	createdAt         time.Time
	maxDelay          time.Duration
	timer             *time.Timer
	attempt           float64
}

func (t *Task) Normalize() {
	if t.Delay == 0 {
		t.Delay = defaultTaskDelay
	}
	if t.ExpFactor == 0 {
		t.ExpFactor = defaultExpFactor
	}
	if t.MaxProcessingTime == 0 {
		t.MaxProcessingTime = defaultTaskMaxProcessingTime
	}
}

func (t *Task) GetDelay() (time.Duration, bool) {
	t.attempt++

	if t.MaxAttempts > 0 && t.attempt >= float64(t.MaxAttempts) {
		return time.Duration(0), false
	}
	if t.MaxProcessingTime > 0 && time.Since(t.createdAt) > t.MaxProcessingTime {
		return time.Duration(0), false
	}

	delay := time.Duration(math.Exp(t.ExpFactor*t.attempt)) * time.Second
	if delay > t.maxDelay {
		delay = t.Delay
	}

	return delay, true
}

func (t *Task) Stop() {
	if t.timer == nil {
		return
	}
	t.timer.Stop()
}
