package devbootstrap

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNormalizeAdminInput(t *testing.T) {
	input, err := NormalizeAdminInput(AdminInput{
		Email:    " Owner@Example.com ",
		Password: "long-enough",
	})
	if err != nil {
		t.Fatalf("expected input to normalize, got %v", err)
	}
	if input.Email != "owner@example.com" {
		t.Fatalf("expected normalized email, got %q", input.Email)
	}
}

func TestNormalizeAdminInputValidation(t *testing.T) {
	tests := []struct {
		name string
		in   AdminInput
		want error
	}{
		{name: "missing email", in: AdminInput{Password: "long-enough"}, want: ErrEmailRequired},
		{name: "invalid email", in: AdminInput{Email: "owner", Password: "long-enough"}, want: ErrEmailInvalid},
		{name: "missing password", in: AdminInput{Email: "owner@example.com"}, want: ErrPasswordRequired},
		{name: "short password", in: AdminInput{Email: "owner@example.com", Password: "short"}, want: ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeAdminInput(tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestHashPasswordUsesBcrypt(t *testing.T) {
	hash, err := HashPassword("long-enough")
	if err != nil {
		t.Fatalf("expected hash, got %v", err)
	}
	if hash == "long-enough" {
		t.Fatalf("expected hashed password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("long-enough")); err != nil {
		t.Fatalf("expected bcrypt hash to verify: %v", err)
	}
}
