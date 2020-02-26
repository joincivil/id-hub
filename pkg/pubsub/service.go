package pubsub

import "github.com/dgrijalva/jwt-go"

// PublisherInterface represents a struct with the nats publish methods
type PublisherInterface interface {
	PublishAdd(*jwt.Token) error
	PublishRevoke(*jwt.Token) error
}
