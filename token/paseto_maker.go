package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(key string) (Maker, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size. must be exactly %d", chacha20poly1305.KeySize)
	}
	return PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(key),
	}, nil
}

func (m PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	token, err := m.paseto.Encrypt(m.symmetricKey, payload, nil)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (m PasetoMaker) VerifyToken(token string) (*Payload, error) {

	payload := &Payload{}
	err := m.paseto.Decrypt(token, m.symmetricKey, payload, nil)
	if err != nil {
		return nil, err
	}

	// It appears Paseto tokens are not validated. We have to manually verify the token has not expired.
    err = payload.Validate()
	if err != nil {
		return nil, err
	}
	return payload, err
}