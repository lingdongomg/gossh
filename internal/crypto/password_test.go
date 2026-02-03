package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct password", password, true},
		{"wrong password", "wrongPassword", false},
		{"empty password", "", false},
		{"similar password", "testPassword123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := VerifyPassword(tt.password, hash)
			if err != nil {
				t.Fatalf("VerifyPassword failed: %v", err)
			}
			if valid != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", valid, tt.want)
			}
		})
	}
}

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantMin  int
		wantMax  int
	}{
		{"empty", "", 0, 0},
		{"short", "abc", 0, 1},
		{"only lowercase", "abcdefgh", 1, 2},
		{"lower + upper", "Abcdefgh", 2, 3},
		{"lower + upper + digit", "Abcdefg1", 3, 4},
		{"all types short", "Abc1!", 2, 3},
		{"all types long", "Abcdefg1!@#", 4, 5},
		{"very strong", "MyP@ssw0rd!2024", 4, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength, _ := PasswordStrength(tt.password)
			if strength < tt.wantMin || strength > tt.wantMax {
				t.Errorf("PasswordStrength(%q) = %d, want between %d and %d", 
					tt.password, strength, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDifferentPasswordsDifferentHashes(t *testing.T) {
	hash1, _ := HashPassword("password1")
	hash2, _ := HashPassword("password2")

	if hash1 == hash2 {
		t.Error("Different passwords should produce different hashes")
	}
}

func TestSamePasswordDifferentHashes(t *testing.T) {
	password := "testPassword"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes (due to random salt)")
	}

	// But both should verify
	valid1, _ := VerifyPassword(password, hash1)
	valid2, _ := VerifyPassword(password, hash2)

	if !valid1 || !valid2 {
		t.Error("Both hashes should verify with the original password")
	}
}
