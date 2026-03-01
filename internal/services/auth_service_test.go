package services

import (
	"os"
	"testing"
)

func TestAuthServiceRegisterLoginAndGetByID(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	svc := NewAuthService(nil) // We ideally need to inject a real connection pool here

	user, err := svc.Register("Alice", "alice@example.com", "strongpass")
	if err != nil {
		t.Fatalf("expected register to succeed, got error: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("expected normalized email, got %s", user.Email)
	}

	_, err = svc.Register("Alice 2", "alice@example.com", "strongpass")
	if err == nil {
		t.Fatalf("expected duplicate email registration to fail")
	}

	loggedIn, err := svc.Login("alice@example.com", "strongpass")
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}
	if loggedIn.ID != user.ID {
		t.Fatalf("expected logged in user id %s, got %s", user.ID, loggedIn.ID)
	}

	_, err = svc.Login("alice@example.com", "wrongpass")
	if err == nil {
		t.Fatalf("expected invalid credentials to fail")
	}

	fetched, err := svc.GetByID(user.ID)
	if err != nil {
		t.Fatalf("expected get by id to succeed, got error: %v", err)
	}
	if fetched.Email != user.Email {
		t.Fatalf("expected fetched email %s, got %s", user.Email, fetched.Email)
	}
}
