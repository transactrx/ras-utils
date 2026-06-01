package rasauth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetToken_BasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected Basic Auth header")
		}
		if user != "test-client" || pass != "test-secret" {
			t.Errorf("unexpected credentials: %s:%s", user, pass)
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != "grant_type=client_credentials" {
			t.Errorf("unexpected body: %s", body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "test-token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	token, err := GetToken(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     server.URL,
		UseBasicAuth: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "test-token" {
		t.Errorf("expected access_token=test-token, got %s", token.AccessToken)
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("expected expires_in=3600, got %d", token.ExpiresIn)
	}
}

func TestGetToken_FormBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); ok {
			t.Error("did not expect Basic Auth header")
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing form: %v", err)
		}

		if r.Form.Get("client_id") != "test-client" {
			t.Errorf("expected client_id=test-client, got %s", r.Form.Get("client_id"))
		}
		if r.Form.Get("client_secret") != "test-secret" {
			t.Errorf("expected client_secret=test-secret, got %s", r.Form.Get("client_secret"))
		}
		if r.Form.Get("scope") != "openid profile" {
			t.Errorf("expected scope=openid profile, got %s", r.Form.Get("scope"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "form-token",
			ExpiresIn:   7200,
			TokenType:   "Bearer",
			IDToken:     "id-token-value",
		})
	}))
	defer server.Close()

	token, err := GetToken(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     server.URL,
		UseBasicAuth: false,
		Scopes:       []string{"openid", "profile"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "form-token" {
		t.Errorf("expected access_token=form-token, got %s", token.AccessToken)
	}
	if token.IDToken != "id-token-value" {
		t.Errorf("expected id_token=id-token-value, got %s", token.IDToken)
	}
}

func TestGetToken_ExtraParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing form: %v", err)
		}

		if r.Form.Get("audience") != "https://api.example.com" {
			t.Errorf("expected audience param, got %s", r.Form.Get("audience"))
		}
		if r.Form.Get("custom_param") != "custom_value" {
			t.Errorf("expected custom_param, got %s", r.Form.Get("custom_param"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{AccessToken: "extra-token"})
	}))
	defer server.Close()

	token, err := GetToken(AuthConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     server.URL,
		ExtraParams: map[string]string{
			"audience":     "https://api.example.com",
			"custom_param": "custom_value",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "extra-token" {
		t.Errorf("expected access_token=extra-token, got %s", token.AccessToken)
	}
}

func TestGetToken_CustomGrantType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing form: %v", err)
		}

		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("expected grant_type=refresh_token, got %s", r.Form.Get("grant_type"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{AccessToken: "refreshed-token"})
	}))
	defer server.Close()

	token, err := GetToken(AuthConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     server.URL,
		GrantType:    "refresh_token",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "refreshed-token" {
		t.Errorf("expected access_token=refreshed-token, got %s", token.AccessToken)
	}
}

func TestGetToken_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer server.Close()

	_, err := GetToken(AuthConfig{
		ClientID:     "bad-client",
		ClientSecret: "bad-secret",
		TokenURL:     server.URL,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected *AuthError, got %T", err)
	}
	if authErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", authErr.StatusCode)
	}
	if authErr.ClientID != "bad-client" {
		t.Errorf("expected client_id=bad-client, got %s", authErr.ClientID)
	}
	if authErr.Message != `{"error":"invalid_client"}` {
		t.Errorf("expected error body in message, got %s", authErr.Message)
	}
}

func TestGetToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	_, err := GetToken(AuthConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     server.URL,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected *AuthError, got %T", err)
	}
	if authErr.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", authErr.StatusCode)
	}
}

func TestGetToken_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{AccessToken: "token"})
	}))
	defer server.Close()

	_, err := GetToken(AuthConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     server.URL,
		Timeout:      50 * time.Millisecond,
	})

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestGetToken_InvalidURL(t *testing.T) {
	_, err := GetToken(AuthConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     "http://localhost:99999/invalid",
	})

	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestGetToken_EmptyTokenURL(t *testing.T) {
	_, err := GetToken(AuthConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     "",
	})

	if err == nil {
		t.Fatal("expected error for empty TokenURL, got nil")
	}
	if err.Error() != "TokenURL required" {
		t.Errorf("expected 'TokenURL required', got %s", err.Error())
	}
}

func TestGetCISToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing form: %v", err)
		}

		if r.Form.Get("scope") != "openid" {
			t.Errorf("expected scope=openid, got %s", r.Form.Get("scope"))
		}
		if r.Form.Get("client_id") != "cis-client" {
			t.Errorf("expected client_id in form body")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "cis-token",
			IDToken:     "cis-id-token",
		})
	}))
	defer server.Close()

	token, err := GetCISToken("cis-client", "cis-secret", server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "cis-token" {
		t.Errorf("expected access_token=cis-token, got %s", token.AccessToken)
	}
}

func TestGetJWTToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected Basic Auth")
		}
		if user != "jwt-client" || pass != "jwt-secret" {
			t.Errorf("unexpected credentials: %s:%s", user, pass)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "jwt-token",
			RefreshToken: "jwt-refresh",
		})
	}))
	defer server.Close()

	token, err := GetJWTToken("jwt-client", "jwt-secret", server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "jwt-token" {
		t.Errorf("expected access_token=jwt-token, got %s", token.AccessToken)
	}
}

func TestAuthError_Error(t *testing.T) {
	err := &AuthError{
		StatusCode: 401,
		Status:     "401 Unauthorized",
		Message:    "invalid credentials",
		TokenURL:   "https://auth.example.com/token",
		ClientID:   "my-client",
	}

	expected := "auth error: invalid credentials (status=401 Unauthorized, client_id=my-client, token_url=https://auth.example.com/token)"
	if err.Error() != expected {
		t.Errorf("unexpected error string:\ngot:  %s\nwant: %s", err.Error(), expected)
	}
}
