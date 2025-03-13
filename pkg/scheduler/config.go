package scheduler

import (
	"time"
)

type Config struct {
	MaxTasksPerSender int
	MaxTaskDelay      time.Duration
}

func (c *Config) normalize() {
	if c.MaxTaskDelay == 0 {
		c.MaxTaskDelay = defaultTaskMaxDelay
	}
}
