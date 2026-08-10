package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type SessionStore interface {
	CreateSession(context.Context, []byte, time.Time) error
	SessionValid(context.Context, []byte, time.Time) (bool, error)
	RevokeSession(context.Context, []byte, time.Time) error
}

type Service struct {
	store        SessionStore
	passwordHash string
	duration     time.Duration
}

func New(store SessionStore, passwordHash string, duration time.Duration) *Service {
	return &Service{store: store, passwordHash: passwordHash, duration: duration}
}

func (s *Service) Login(ctx context.Context, password string, now time.Time) (string, time.Time, error) {
	valid, err := VerifyPassword(password, s.passwordHash)
	if err != nil || !valid {
		return "", time.Time{}, ErrInvalidCredentials
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expiresAt := now.Add(s.duration)
	if err := s.store.CreateSession(ctx, tokenHash(token), expiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("persist session: %w", err)
	}
	return token, expiresAt, nil
}

func (s *Service) Valid(ctx context.Context, token string, now time.Time) (bool, error) {
	if token == "" {
		return false, nil
	}
	return s.store.SessionValid(ctx, tokenHash(token), now)
}

func (s *Service) Logout(ctx context.Context, token string, now time.Time) error {
	if token == "" {
		return nil
	}
	return s.store.RevokeSession(ctx, tokenHash(token), now)
}

func tokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, errors.New("invalid Argon2id hash")
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
