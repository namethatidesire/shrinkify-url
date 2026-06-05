package shortener

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
)

const codeLength = 6

var ErrInvalidURL = errors.New("invalid URL")

// Returns a random URL-safe base64 code of length codeLength.
func GenerateCode() (string, error) {
	b := make([]byte, codeLength)
	_, err := rand.Read(b) // technically: never returns an error.
	if err != nil {
		return "", err
	}
	code := base64.URLEncoding.EncodeToString(b)
	return code[:codeLength], nil
}

// Minimal check that a string url starts with http:// or https://.
func ValidateURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ErrInvalidURL
	}
	return nil
}