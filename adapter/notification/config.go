package notification

import "time"

type Config struct {
	BaseURL  string
	Timeout  time.Duration
	Priority string
	PoolSize int
}
