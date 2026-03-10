package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/pquerna/otp/totp"
	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type nopCloser struct {
	*bytes.Buffer
}

func (nopCloser) Close() error { return nil }

// GenerateTOTP creates a new TOTP secret and returns the secret string and a QR code as base64 PNG.
func GenerateTOTP(username, issuer string) (secret, qrBase64 string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate totp: %w", err)
	}

	secret = key.Secret()
	uri := key.URL()

	qrc, qrErr := qrcode.NewWith(uri)
	if qrErr != nil {
		return secret, "", nil
	}

	var buf bytes.Buffer
	wr := standard.NewWithWriter(nopCloser{&buf})
	if saveErr := qrc.Save(wr); saveErr != nil {
		return secret, "", nil
	}

	qrBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
	return secret, qrBase64, nil
}

// ValidateTOTP checks a TOTP code against a secret.
func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
