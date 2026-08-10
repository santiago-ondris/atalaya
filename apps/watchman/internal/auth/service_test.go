package auth

import (
	"context"
	"testing"
	"time"
)

const testHash = "$argon2id$v=19$m=65536,t=3,p=2$YXRhbGF5YS1kZXYtc2FsdA$Ixko1/mnbcEDH1CiWYhtyGiUELM98O68mONv/YEIFo0"

type memoryStore struct {
	hash    []byte
	expires time.Time
	revoked bool
}

func (store *memoryStore) CreateSession(_ context.Context, hash []byte, expires time.Time) error {
	store.hash = hash
	store.expires = expires
	return nil
}
func (store *memoryStore) SessionValid(_ context.Context, hash []byte, now time.Time) (bool, error) {
	return string(hash) == string(store.hash) && now.Before(store.expires) && !store.revoked, nil
}
func (store *memoryStore) RevokeSession(_ context.Context, hash []byte, _ time.Time) error {
	if string(hash) == string(store.hash) {
		store.revoked = true
	}
	return nil
}

func TestLoginCreatesRevocableOpaqueSession(t *testing.T) {
	store := &memoryStore{}
	service := New(store, testHash, time.Hour)
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	token, expires, err := service.Login(context.Background(), "atalaya_local", now)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || string(store.hash) == token {
		t.Fatal("expected an opaque token stored only as a hash")
	}
	if !expires.Equal(now.Add(time.Hour)) {
		t.Fatalf("unexpected expiry %v", expires)
	}
	valid, err := service.Valid(context.Background(), token, now)
	if err != nil || !valid {
		t.Fatalf("expected valid session: %v", err)
	}
	if err := service.Logout(context.Background(), token, now); err != nil {
		t.Fatal(err)
	}
	valid, _ = service.Valid(context.Background(), token, now)
	if valid {
		t.Fatal("expected revoked session to be invalid")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	_, _, err := New(&memoryStore{}, testHash, time.Hour).Login(context.Background(), "wrong", time.Now())
	if err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
