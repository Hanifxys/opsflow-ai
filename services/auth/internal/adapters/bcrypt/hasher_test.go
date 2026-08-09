package bcrypt_test

import (
	"testing"

	"github.com/opsflow/auth-service/internal/adapters/bcrypt"
)

func TestBcryptHasher(t *testing.T) {
	hasher := bcrypt.New(4) // low cost for fast test
	password := "my-secret-password"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !hasher.Compare(password, hash) {
		t.Error("expected password to match hash")
	}

	if hasher.Compare("wrong-password", hash) {
		t.Error("expected wrong password to fail match")
	}
}
