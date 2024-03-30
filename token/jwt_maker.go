package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {

	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d character", minSecretKeySize)
	}

	return &JWTMaker{secretKey}, nil
}
func (j *JWTMaker) CreateToken(username string, role string, duration time.Duration) (string, *Payload, error) {

	payload, err := NewPayload(username, role, duration)

	if err != nil {
		return "", payload, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	jwtToken, err := token.SignedString([]byte(j.secretKey))

	if err != nil {
		return "", payload, err
	}
	return jwtToken, payload, err
}

func (j *JWTMaker) VerifyToken(tokenID string) (*Payload, error) {

	// For checking SigningMethodHMAC is same or not base on previous usage in CreateToken func
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {

			return nil, ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(tokenID, &Payload{}, keyFunc, jwt.WithLeeway(5*time.Second))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}

		return nil, ErrInvalidToken
	}

	if payload, ok := jwtToken.Claims.(*Payload); !ok {
		return nil, jwt.ErrTokenInvalidClaims
	} else {
		return payload, nil
	}

}
