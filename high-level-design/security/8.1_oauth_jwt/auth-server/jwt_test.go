package main

import (
	"errors"
	"testing"
	"time"
)

var testSecret = []byte("test-secret-key")

func TestIssueAndValidate(t *testing.T) {
	token, err := IssueToken(testSecret, "user123", "admin", time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := ValidateToken(testSecret, token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.Sub != "user123" {
		t.Errorf("sub: expected user123, got %q", claims.Sub)
	}
	if claims.Role != "admin" {
		t.Errorf("role: expected admin, got %q", claims.Role)
	}
}

func TestExpiredToken(t *testing.T) {
	token, _ := IssueToken(testSecret, "user1", "user", -time.Second)
	_, err := ValidateToken(testSecret, token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestTamperedSignature(t *testing.T) {
	token, _ := IssueToken(testSecret, "user1", "user", time.Hour)
	tampered := token[:len(token)-4] + "XXXX"
	_, err := ValidateToken(testSecret, tampered)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("expected ErrTokenInvalid for tampered token, got %v", err)
	}
}

func TestWrongSecret(t *testing.T) {
	token, _ := IssueToken(testSecret, "user1", "user", time.Hour)
	_, err := ValidateToken([]byte("wrong-secret"), token)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("expected ErrTokenInvalid for wrong secret, got %v", err)
	}
}

func TestMalformedToken(t *testing.T) {
	_, err := ValidateToken(testSecret, "not.a.valid.jwt.token")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}
