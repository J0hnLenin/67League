package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/J0hnLenin/67League/internal/application/ports"
	"github.com/J0hnLenin/67League/internal/domain/shared"
)

const (
	subjectPlayer = "player"
	subjectAdmin  = "admin"
)

// JWTIssuer реализует ports.TokenIssuer через JWT.
type JWTIssuer struct {
	secret    []byte
	playerTTL time.Duration
	adminTTL  time.Duration
}

type JWTConfig struct {
	Secret    string
	PlayerTTL time.Duration
	AdminTTL  time.Duration
}

func NewJWTIssuer(cfg JWTConfig) *JWTIssuer {
	if cfg.PlayerTTL <= 0 {
		cfg.PlayerTTL = 30 * 24 * time.Hour
	}
	if cfg.AdminTTL <= 0 {
		cfg.AdminTTL = 12 * time.Hour
	}
	return &JWTIssuer{
		secret:    []byte(cfg.Secret),
		playerTTL: cfg.PlayerTTL,
		adminTTL:  cfg.AdminTTL,
	}
}

type jwtClaims struct {
	Subject string `json:"sub"`
	UserID  string `json:"uid"`
	jwt.RegisteredClaims
}

func (j *JWTIssuer) IssuePlayer(id shared.PlayerID) (string, error) {
	return j.issue(subjectPlayer, id.String(), j.playerTTL)
}

func (j *JWTIssuer) IssueAdmin(id shared.AdminID) (string, error) {
	return j.issue(subjectAdmin, id.String(), j.adminTTL)
}

func (j *JWTIssuer) issue(subject, uid string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwtClaims{
		Subject: subject,
		UserID:  uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

var errInvalidToken = errors.New("invalid token")

func (j *JWTIssuer) Parse(token string) (*ports.AuthClaims, error) {
	var claims jwtClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errInvalidToken
		}
		return j.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errInvalidToken
	}

	subj := ports.TokenSubject(claims.Subject)
	switch subj {
	case ports.SubjectPlayer, ports.SubjectAdmin:
	default:
		return nil, errInvalidToken
	}

	return &ports.AuthClaims{
		Subject:   subj,
		UserID:    claims.UserID,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
