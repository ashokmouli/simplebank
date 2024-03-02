package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrExpiredToken = errors.New("token has expired")

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("uuid creation failed: %w", err)
	}
	return &Payload{
		ID:        uuid,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}

func (p *Payload) Validate() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
