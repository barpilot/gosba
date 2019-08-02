package crypto

// Config represents configuration options for the global codec
type Config struct {
	EncryptionScheme string
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return Config{}
}
