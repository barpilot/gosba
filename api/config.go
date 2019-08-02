package api

// Config represents configuration options for the API server
type Config struct {
	Port        int
	TLSCertPath string
	TLSKeyPath  string
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return Config{Port: 8080}
}
