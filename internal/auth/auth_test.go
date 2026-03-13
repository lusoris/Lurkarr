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

func TestContainsGroup(t *testing.T) {
	tests := []struct {
		name   string
		groups []string
		target string
		want   bool
	}{
		{"exact match", []string{"admin", "users"}, "admin", true},
		{"case insensitive", []string{"Admin", "Users"}, "admin", true},
		{"not found", []string{"users", "editors"}, "admin", false},
		{"empty groups", nil, "admin", false},
		{"empty target", []string{"admin"}, "", false},
		{"multiple groups with match", []string{"group1", "group2", "lurkarr-admins"}, "lurkarr-admins", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsGroup(tt.groups, tt.target); got != tt.want {
				t.Errorf("containsGroup(%v, %q) = %v, want %v", tt.groups, tt.target, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Recovery code tests
// =============================================================================

func TestGenerateRecoveryCodes(t *testing.T) {
	plain, hashed, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plain) != 10 {
		t.Fatalf("expected 10 plain codes, got %d", len(plain))
	}
	if len(hashed) != 10 {
		t.Fatalf("expected 10 hashed codes, got %d", len(hashed))
	}
	// Each plain code should be 9 chars (4-4 with dash)
	for i, code := range plain {
		if len(code) != 9 {
			t.Errorf("code %d: expected 9 chars, got %d (%q)", i, len(code), code)
		}
		if code[4] != '-' {
			t.Errorf("code %d: expected dash at position 4, got %q", i, code)
		}
	}
}

func TestGenerateRecoveryCodes_Uniqueness(t *testing.T) {
	plain, _, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen := make(map[string]bool, len(plain))
	for _, code := range plain {
		if seen[code] {
			t.Fatalf("duplicate recovery code: %s", code)
		}
		seen[code] = true
	}
}

func TestValidateRecoveryCode_Valid(t *testing.T) {
	plain, hashed, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Each plain code should match its corresponding hash
	for i, code := range plain {
		idx := ValidateRecoveryCode(code, hashed)
		if idx != i {
			t.Errorf("expected code %d to match at index %d, got %d", i, i, idx)
		}
	}
}

func TestValidateRecoveryCode_Invalid(t *testing.T) {
	_, hashed, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	idx := ValidateRecoveryCode("0000-0000", hashed)
	if idx != -1 {
		t.Fatalf("expected -1 for invalid code, got %d", idx)
	}
}

func TestValidateRecoveryCode_EmptyHashes(t *testing.T) {
	idx := ValidateRecoveryCode("abcd-ef01", nil)
	if idx != -1 {
		t.Fatalf("expected -1 for empty hashes, got %d", idx)
	}
}

func TestValidateRecoveryCode_EmptyCode(t *testing.T) {
	_, hashed, err := GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	idx := ValidateRecoveryCode("", hashed)
	if idx != -1 {
		t.Fatalf("expected -1 for empty code, got %d", idx)
	}
}
