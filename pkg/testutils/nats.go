package testutils

import (
	"github.com/dgrijalva/jwt-go"
)

type FakeNatsService struct{}

func (*FakeNatsService) PublishAdd(token *jwt.Token) error {
	return nil
}

func (*FakeNatsService) PublishRevoke(token *jwt.Token) error {
	return nil
}
