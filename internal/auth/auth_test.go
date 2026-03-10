package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "testpassword123" {
		t.Fatal("HashPassword returned plaintext password")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("password")
	h2, _ := HashPassword("password")
	if h1 == h2 {
		t.Fatal("identical passwords should produce different bcrypt hashes (salt)")
	}
}

func TestCheckPassword(t *testing.T) {
	hash, _ := HashPassword("correctpassword")

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct password", "correctpassword", true},
		{"wrong password", "wrongpassword", false},
		{"empty password", "", false},
		{"similar password", "correctpasswor", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPassword(tt.password, hash)
			if got != tt.want {
				t.Errorf("CheckPassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestCheckPasswordInvalidHash(t *testing.T) {
	if CheckPassword("password", "nothash") {
		t.Fatal("CheckPassword should return false for invalid hash")
	}
}

func TestGenerateSecretKey(t *testing.T) {
	key, err := GenerateSecretKey(32)
	if err != nil {
		t.Fatalf("GenerateSecretKey returned error: %v", err)
	}
	if len(key) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(key))
	}
}

func TestGenerateSecretKeyUniqueness(t *testing.T) {
	k1, _ := GenerateSecretKey(32)
	k2, _ := GenerateSecretKey(32)
	if k1 == k2 {
		t.Fatal("two generated keys should not be identical")
	}
}

func TestGenerateSecretKeyLengths(t *testing.T) {
	tests := []struct {
		bytes  int
		hexLen int
	}{
		{16, 32},
		{32, 64},
		{64, 128},
		{1, 2},
	}
	for _, tt := range tests {
		key, err := GenerateSecretKey(tt.bytes)
		if err != nil {
			t.Fatalf("GenerateSecretKey(%d) error: %v", tt.bytes, err)
		}
		if len(key) != tt.hexLen {
			t.Errorf("GenerateSecretKey(%d) = %d chars, want %d", tt.bytes, len(key), tt.hexLen)
		}
	}
}

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashPassword("benchmarkpassword")
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	hash, _ := HashPassword("benchmarkpassword")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPassword("benchmarkpassword", hash)
	}
}
