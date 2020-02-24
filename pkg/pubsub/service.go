package pubsub

import "github.com/dgrijalva/jwt-go"

// Interface represents a struct with the nats publish methods
type Interface interface {
	PublishAdd(*jwt.Token) error
	PublishRevoke(*jwt.Token) error
}
