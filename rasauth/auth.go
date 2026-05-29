// Package rasauth provides OAuth2 client credentials token acquisition.
package rasauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/transactrx/ras-utils/rashttp"
)

// AuthConfig configures OAuth2 client credentials token requests.
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	GrantType    string            // defaults to "client_credentials"
	Scopes       []string          // optional scopes to request
	UseBasicAuth bool              // true = Basic Auth header, false = credentials in form body
	Timeout      time.Duration     // defaults to 10s
	ExtraParams  map[string]string // additional form parameters
}

// TokenResponse represents a standard OAuth2 token response.
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token,omitempty"`
	NotBeforePolicy  int64  `json:"not-before-policy,omitempty"`
	SessionState     string `json:"session_state,omitempty"`
	Scope            string `json:"scope,omitempty"`
}

// AuthError represents an authentication failure.
type AuthError struct {
	StatusCode int
	Status     string
	Message    string
	TokenURL   string
	ClientID   string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error: %s (status=%s, client_id=%s, token_url=%s)",
		e.Message, e.Status, e.ClientID, e.TokenURL)
}

// GetToken acquires an OAuth2 token using client credentials flow.
func GetToken(config AuthConfig) (TokenResponse, error) {
	// Apply defaults
	if config.GrantType == "" {
		config.GrantType = "client_credentials"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// Build form data
	data := url.Values{}
	data.Set("grant_type", config.GrantType)

	if !config.UseBasicAuth {
		data.Set("client_id", config.ClientID)
		data.Set("client_secret", config.ClientSecret)
	}

	if len(config.Scopes) > 0 {
		data.Set("scope", strings.Join(config.Scopes, " "))
	}

	for k, v := range config.ExtraParams {
		data.Set(k, v)
	}

	encodedData := data.Encode()

	req, err := http.NewRequest(http.MethodPost, config.TokenURL, strings.NewReader(encodedData))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(encodedData)))

	if config.UseBasicAuth {
		req.SetBasicAuth(config.ClientID, config.ClientSecret)
	}

	client := rashttp.NewHttpClient(config.Timeout)
	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TokenResponse{}, &AuthError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Message:    "token request failed",
			TokenURL:   config.TokenURL,
			ClientID:   config.ClientID,
		}
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return TokenResponse{}, &AuthError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Message:    fmt.Sprintf("decoding response: %v", err),
			TokenURL:   config.TokenURL,
			ClientID:   config.ClientID,
		}
	}

	return token, nil
}

// GetCISToken fetches a token with credentials in the form body and openid scope.
func GetCISToken(clientId string, clientSecret string, tokenUrl string) (TokenResponse, error) {
	config := AuthConfig{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     tokenUrl,
		UseBasicAuth: false,
		Scopes:       []string{"openid"},
		Timeout:      10 * time.Second,
	}
	return GetToken(config)
}

// GetJWTToken fetches a token using Basic Auth header.
func GetJWTToken(clientId string, clientSecret string, tokenUrl string) (TokenResponse, error) {
	config := AuthConfig{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     tokenUrl,
		UseBasicAuth: true,
		Timeout:      10 * time.Second,
	}

	return GetToken(config)
}
