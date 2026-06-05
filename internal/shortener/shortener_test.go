package shortener

import (
	"testing"
)

func TestGenerateCode_Length(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != codeLength {
		t.Errorf("expected length %d, got %d", codeLength, len(code))
	}
}

func TestGenerateCode_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if seen[code] {
			t.Fatalf("duplicate code generated: %s", code)
		}
		seen[code] = true
	}
}

func TestValidateURL_Valid(t *testing.T) {
	valid := []string{
		"https://www.google.com/",
		"http://www.example.org/path?q=1&test=true",
	}
	for _, url := range valid {
		if err := ValidateURL(url); err != nil {
			t.Errorf("expected valid, got error for %q: %v", url, err)
		}
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	invalid := []string{
		"www.google.com/",
		"example.org/path?q=1&test=true",
		"",
		"ftp://realwebsite.net",
		"127.0.0.1",
		"javascript:alert(1)",
	}
	for _, url := range invalid {
		if err := ValidateURL(url); err == nil {
			t.Errorf("expected error, got valid for %q: %v", url, err)
		}
	}
}
