package api

// BasicAuthConfig represents details such as username and password that will
// be used to secure the broker using basic auth
type BasicAuth interface {
	GetUsername() string
	GetPassword() string
}

type BasicAuthConfig struct {
	Username   string
	CFUsername string
	Password   string
	CFPassword string
}

func (b BasicAuthConfig) GetUsername() string {
	return b.Username
}

func (b BasicAuthConfig) GetPassword() string {
	return b.Password
}
