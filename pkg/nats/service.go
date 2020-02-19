package nats

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/joincivil/id-hub/pkg/didjwt"
	stan "github.com/nats-io/stan.go"
)

type NatsInterface interface {
	PublishAdd(*jwt.Token) error
	PublishRevoke(*jwt.Token) error
}

type Service struct {
	NatsClient    stan.Conn
	SubjectPrefix string
}

func NewService(sc stan.Conn, sp string) *Service {
	return &Service{
		NatsClient:    sc,
		SubjectPrefix: sp,
	}
}

func (s *Service) PublishAdd(token *jwt.Token) error {
	claims, ok := token.Claims.(*didjwt.VCClaimsJWT)
	if !ok {
		return errors.New("invalid claims type")
	}

	err := s.NatsClient.Publish(fmt.Sprintf("%s.%s.add", s.SubjectPrefix, claims.Issuer), []byte(token.Raw))
	if err != nil {
		return errors.Wrap(err, "PublishAdd failed to publish for the DID")
	}

	// TODO(walfly): Only publish to the public channel when claim is public

	err = s.NatsClient.Publish(fmt.Sprintf("%s.public.add", s.SubjectPrefix), []byte(token.Raw))
	if err != nil {
		return errors.Wrap(err, "PublishAdd failed to publish to the public channel")
	}

	return nil
}

func (s *Service) PublishRevoke(token *jwt.Token) error {
	claims, ok := token.Claims.(*didjwt.VCClaimsJWT)
	if !ok {
		return errors.New("invalid claims type")
	}

	err := s.NatsClient.Publish(fmt.Sprintf("%s.%s.revoke", s.SubjectPrefix, claims.Issuer), []byte(token.Raw))
	if err != nil {
		return errors.Wrap(err, "PublishRevoke failed to publish for the DID")
	}

	// TODO(walfly): Only publish to the public channel when claim is public

	err = s.NatsClient.Publish(fmt.Sprintf("%s.public.revoke", s.SubjectPrefix), []byte(token.Raw))
	if err != nil {
		return errors.Wrap(err, "PublishRevoke failed to publish to the public channel")
	}

	return nil
}
