package strutil_test

import (
	"strings"
	"testing"

	"github.com/hanjo-erp/hanjo/pkg/strutil"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Sepatu Olahraga Pria #1!", "sepatu-olahraga-pria-1"},
		{"  Hello   World  ", "hello-world"},
		{"Golang Clean Architecture & DDD", "golang-clean-architecture-ddd"},
		{"", ""},
		{"---test---", "test"},
	}

	for _, tc := range cases {
		res := strutil.Slugify(tc.input)
		if res != tc.expected {
			t.Errorf("Slugify(%q) = %q, expected %q", tc.input, res, tc.expected)
		}
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"admin@hanjo-erp.com", "a***n@hanjo-erp.com"},
		{"ab@domain.com", "a***@domain.com"},
		{"invalid-email", "invalid-email"},
	}

	for _, tc := range cases {
		res := strutil.MaskEmail(tc.input)
		if res != tc.expected {
			t.Errorf("MaskEmail(%q) = %q, expected %q", tc.input, res, tc.expected)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	phone := "081234567890"
	masked := strutil.MaskPhone(phone)
	if !strings.HasPrefix(masked, "0812") || !strings.HasSuffix(masked, "7890") || !strings.Contains(masked, "****") {
		t.Errorf("unexpected masked phone: %s", masked)
	}

	short := "12345"
	if strutil.MaskPhone(short) != "12345" {
		t.Errorf("expected short phone unchanged")
	}
}

func TestRandomNumeric(t *testing.T) {
	otp, err := strutil.RandomNumeric(6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(otp) != 6 {
		t.Errorf("expected length 6, got %d", len(otp))
	}
	for _, ch := range otp {
		if ch < '0' || ch > '9' {
			t.Errorf("expected digit, got character %c", ch)
		}
	}

	empty, _ := strutil.RandomNumeric(0)
	if empty != "" {
		t.Errorf("expected empty string for length 0")
	}
}

func TestRandomAlphanumeric(t *testing.T) {
	str, err := strutil.RandomAlphanumeric(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(str) != 16 {
		t.Errorf("expected length 16, got %d", len(str))
	}
}

func TestToCamelCase(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"created_at", "createdAt"},
		{"user-profile-id", "userProfileId"},
		{"alreadyCamel", "alreadycamel"},
		{"", ""},
	}

	for _, tc := range cases {
		res := strutil.ToCamelCase(tc.input)
		if res != tc.expected {
			t.Errorf("ToCamelCase(%q) = %q, expected %q", tc.input, res, tc.expected)
		}
	}
}
