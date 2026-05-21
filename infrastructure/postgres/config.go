package postgres

type Config struct {
	Host          string
	Port          int
	Username      string
	Password      string
	Database      string
	MigrationPath string
}

var cfg = &Config{
	Host:          "localhost",
	Port:          5432,
	Username:      "taskforge",
	Password:      "taskforge",
	Database:      "taskforge",
	MigrationPath: "file://migrations",
}

func SetConfig(c Config) {
	cfg = &c
}

func (c *Config) Validation() error {
	if c.Host == "" || c.Database == "" || c.Username == "" {
		return ErrInvalidPostgresConfig
	}
	if c.Port <= 0 {
		return ErrInvalidPostgresConfig
	}
	return nil
}
