package testutils

import (
	"github.com/dgrijalva/jwt-go"
)

// FakePubSubService stubs out a pub sub service
type FakePubSubService struct{}

// PublishAdd does nothing
func (*FakePubSubService) PublishAdd(token *jwt.Token) error {
	return nil
}

// PublishRevoke does nothing
func (*FakePubSubService) PublishRevoke(token *jwt.Token) error {
	return nil
}
