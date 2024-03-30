package token

import "time"

// Maker interface aim to switch easily between JWT and Paseto token
type Maker interface {
	CreateToken(username string, role string, duration time.Duration) (string, *Payload, error)

	VerifyToken(tokenID string) (*Payload, error)
}
