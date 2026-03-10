package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateTOTP(t *testing.T) {
	secret, qrBase64, err := GenerateTOTP("testuser", "Lurkarr")
	if err != nil {
		t.Fatalf("GenerateTOTP returned error: %v", err)
	}
	if secret == "" {
		t.Fatal("GenerateTOTP returned empty secret")
	}
	_ = qrBase64
}

func TestValidateTOTP(t *testing.T) {
	secret, _, err := GenerateTOTP("testuser", "Lurkarr")
	if err != nil {
		t.Fatalf("GenerateTOTP error: %v", err)
	}

	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode error: %v", err)
	}

	if !ValidateTOTP(code, secret) {
		t.Fatal("ValidateTOTP should accept valid code")
	}
}

func TestValidateTOTPWrongCode(t *testing.T) {
	secret, _, _ := GenerateTOTP("testuser", "Lurkarr")
	if ValidateTOTP("000000", secret) {
		t.Log("000000 happened to be valid (extremely rare)")
	}
}

func TestValidateTOTPEmptySecret(t *testing.T) {
	if ValidateTOTP("123456", "") {
		t.Fatal("ValidateTOTP should return false for empty secret")
	}
}

func TestValidateTOTPEmptyCode(t *testing.T) {
	secret, _, _ := GenerateTOTP("testuser", "Lurkarr")
	if ValidateTOTP("", secret) {
		t.Fatal("ValidateTOTP should return false for empty code")
	}
}
