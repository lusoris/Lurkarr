package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/pquerna/otp/totp"
	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// nopCloseBuffer wraps bytes.Buffer to satisfy io.WriteCloser.
type nopCloseBuffer struct {
	bytes.Buffer
}

func (nopCloseBuffer) Close() error { return nil }

// GenerateTOTP creates a new TOTP secret and returns the secret string and a QR code as base64 JPEG.
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

	qrc, err := qrcode.NewWith(uri)
	if err != nil {
		return "", "", fmt.Errorf("create qr code: %w", err)
	}

	var buf nopCloseBuffer
	wr := standard.NewWithWriter(&buf)
	if saveErr := qrc.Save(wr); saveErr != nil {
		// Fallback: return secret without QR if rendering fails
		return secret, "", nil //nolint:nilerr // intentional fallback
	}

	qrBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
	return secret, qrBase64, nil
}

// ValidateTOTP verifies a TOTP code against a secret.
func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
