package token

import (
	"github.com/HL/meta-bank/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestJWTMaker(t *testing.T) {

	maker, err := NewJWTMaker(util.RandomString(32))

	require.NoError(t, err)

	username := util.RandomOwner()
	role := util.DEPOSITOR
	duration := time.Minute

	issuedAt := time.Now()
	expiresAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, role, duration)

	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotEmpty(t, token)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)

	require.Equal(t, username, payload.Username)
	require.Equal(t, role, payload.Role)

	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiresAt, payload.ExpiredAt, time.Second)

}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))

	require.NoError(t, err)

	username := util.RandomOwner()
	role := util.DEPOSITOR
	duration := -time.Minute

	token, payload, err := maker.CreateToken(username, role, duration)

	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotEmpty(t, token)

	payload, err = maker.VerifyToken(token)

	require.Error(t, err)
	require.EqualError(t, err, ErrTokenExpired.Error())
	require.Nil(t, payload)

}

func TestInvalidJWTTokenAlgNone(t *testing.T) {

	payload, err := NewPayload(util.RandomOwner(), util.DEPOSITOR, time.Minute)

	require.NoError(t, err)

	claims := &jwt.RegisteredClaims{
		ID:        payload.ID.String(),
		Issuer:    payload.Username,
		Subject:   payload.Role,
		ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
		IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTMaker(util.RandomString(33))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	require.Nil(t, payload)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())

}
