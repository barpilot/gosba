package aes256

// Config represents configuration options for the AES256-based implementation
// of the Crypto interface
type Config struct {
	Key string
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return Config{}
}
