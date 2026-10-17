package strutil

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	multipleDashRegex    = regexp.MustCompile(`-+`)
)

// Slugify converts a string into a URL-friendly lowercase slug.
// Example: "Sepatu Olahraga Pria #1!" -> "sepatu-olahraga-pria-1"
func Slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}

	// Remove non-alphanumeric runes
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	// Deduplicate dashes
	s = multipleDashRegex.ReplaceAllString(s, "-")
	// Trim trailing/leading dashes
	s = strings.Trim(s, "-")

	return s
}

// MaskEmail masks the local part of an email address for privacy.
// Example: "john.doe@example.com" -> "j***e@example.com"
func MaskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local[:1] + "***@" + domain
	}

	return string(local[0]) + "***" + string(local[len(local)-1]) + "@" + domain
}

// MaskPhone masks middle digits of a phone number.
// Example: "081234567890" -> "0812****7890"
func MaskPhone(phone string) string {
	cleaned := strings.TrimSpace(phone)
	if len(cleaned) <= 6 {
		return cleaned
	}

	prefixLen := 4
	suffixLen := 4
	if len(cleaned) < 10 {
		prefixLen = 2
		suffixLen = 2
	}

	prefix := cleaned[:prefixLen]
	suffix := cleaned[len(cleaned)-suffixLen:]
	return prefix + "****" + suffix
}

const (
	numericChars      = "0123456789"
	alphanumericChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// RandomNumeric generates a cryptographically secure random numeric string of given length.
// Useful for OTP generation.
func RandomNumeric(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}
	bytes := make([]byte, length)
	max := big.NewInt(int64(len(numericChars)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		bytes[i] = numericChars[n.Int64()]
	}

	return string(bytes), nil
}

// RandomAlphanumeric generates a cryptographically secure random string containing letters and numbers.
func RandomAlphanumeric(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}
	bytes := make([]byte, length)
	max := big.NewInt(int64(len(alphanumericChars)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		bytes[i] = alphanumericChars[n.Int64()]
	}

	return string(bytes), nil
}

// ToCamelCase converts snake_case or kebab-case to camelCase.
func ToCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || unicode.IsSpace(r)
	})

	if len(parts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(strings.ToLower(parts[0]))

	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		runes := []rune(p)
		sb.WriteRune(unicode.ToUpper(runes[0]))
		sb.WriteString(strings.ToLower(string(runes[1:])))
	}

	return sb.String()
}
