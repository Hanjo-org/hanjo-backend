package domain

import "time"

// TokenPair represents an issued access and refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"` // seconds
	ExpiresAt    time.Time `json:"expires_at"`
}
