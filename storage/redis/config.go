package redis

// Config represents configuration options for the Redis-based implementation
// of the Store interface
type Config struct {
	RedisHost      string
	RedisPort      int
	RedisPassword  string
	RedisDB        int
	RedisEnableTLS bool
	RedisPrefix    string
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return Config{RedisPort: 6379}
}
