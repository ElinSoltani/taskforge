package redis

type Config struct {
	Addr          string
	Password      string
	DB            int
	Stream        string
	ConsumerGroup string
	ConsumerName  string
	BlockTimeout  int // milliseconds for XREADGROUP block
}

var cfg = &Config{
	Addr:          "localhost:6379",
	Stream:        "taskforge:queue:normal",
	ConsumerGroup: "taskforge-workers",
	ConsumerName:  "worker-1",
	BlockTimeout:  2000,
}

func SetConfig(c Config) {
	cfg = &c
}

func (c *Config) Validation() error {
	if c.Addr == "" || c.Stream == "" || c.ConsumerGroup == "" {
		return ErrInvalidRedisConfig
	}
	return nil
}
