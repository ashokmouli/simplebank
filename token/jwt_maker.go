package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct {
	secretKey string
}

const minSecretKeySize = 30

var ErrInvalidToken = errors.New("token is invalid")

func NewJWTMaker(secret string) (Maker, error) {
	if len(secret) < minSecretKeySize {
		return nil, fmt.Errorf("keylength should be at least 30 characters")
	}
	return &JWTMaker{
		secretKey: secret,
	}, nil
}

// Wraps a Payload object and provides a claims interface.
type RegClaims struct {
	Payload
}

func (c RegClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.Payload.ExpiredAt), nil
}

// GetNotBefore implements the Claims interface.
func (c RegClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

// GetIssuedAt implements the Claims interface.
func (c RegClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.Payload.IssuedAt), nil
}

// GetIssuer implements the Claims interface.
func (c RegClaims) GetIssuer() (string, error) {
	return "none", nil
}

// GetAudience implements the Claims interface.
func (c RegClaims) GetAudience() (jwt.ClaimStrings, error) {
	return []string{"none"}, nil
}

// GetSubject implements the Claims interface.
func (c RegClaims) GetSubject() (string, error) {
	return "none", nil
}

func (j *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, RegClaims{Payload: *payload})
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	return token, err

}
func (j *JWTMaker) VerifyToken(token string) (*Payload, error) {
	
	p := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	registeredClaims := RegClaims{}

	jwtToken, err := p.ParseWithClaims(token, registeredClaims, func(token *jwt.Token) (interface{}, error) {
		// These checks are not required as the WithValidMethods will validate the signing method.
		// However, we will keep it for a second level of check.
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	
	registerdClaims, ok := jwtToken.Claims.(*RegClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return &registerdClaims.Payload, nil
}
