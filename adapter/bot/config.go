package bot

type Config struct {
	Token         string
	ChatId        []int64
	UpdateTimeout int
	PoolSize      int
}
