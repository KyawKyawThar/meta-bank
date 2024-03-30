package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
)

type Payload struct {
	jwt.RegisteredClaims
}

func NewPayload(username string, role string, duration time.Duration) (*Payload, error) {
	tokenUID, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}
	payload := &Payload{

		jwt.RegisteredClaims{
			ID:        tokenUID.String(),
			Issuer:    username,
			Subject:   role,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return payload, nil
}
