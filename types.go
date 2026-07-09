package gigya

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNetworkError     = errors.New("gigya: network error")
	ErrNotAuthenticated = errors.New("gigya: not authenticated")
)

type GigyaResponse struct {
	CallID       string `json:"callId"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	ErrorDetails string `json:"errorDetails,omitempty"`
	StatusCode   int    `json:"statusCode"`
	StatusReason string `json:"statusReason"`
	Time         string `json:"time"`
}

func (r GigyaResponse) IsError() bool {
	return r.ErrorCode != 0
}

type APIError struct {
	Response GigyaResponse
	Raw 	 []byte
}

func (e *APIError) Error() string {
	if e.Response.ErrorDetails != "" {
		return fmt.Sprintf("gigya: %s (errorCode=%d, details=%s)", e.Response.ErrorMessage, e.Response.ErrorCode, e.Response.ErrorDetails)
	}
	return fmt.Sprintf("gigya: %s (errorCode=%d)", e.Response.ErrorMessage, e.Response.ErrorCode)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Account struct {
	UID           string         `json:"UID"`
	IsRegistered  bool           `json:"isRegistered"`
	IsActive      bool           `json:"isActive"`
	IsVerified    bool           `json:"isVerified"`
	Created       string         `json:"created"`
	LastLogin     string         `json:"lastLogin"`
	LastUpdated   string         `json:"lastUpdated"`
	LoginProvider string         `json:"loginProvider,omitempty"`
	Emails        *AccountEmails `json:"emails,omitempty"`
	Profile       map[string]any `json:"profile,omitempty"`
	Data          map[string]any `json:"data,omitempty"`
	Preferences   map[string]any `json:"preferences,omitempty"`
	LoginIDs      *LoginIDs      `json:"loginIDs,omitempty"`
}

type AccountEmails struct {
	Verified   []string `json:"verified,omitempty"`
	Unverified []string `json:"unverified,omitempty"`
}

type LoginIDs struct {
	Username string   `json:"username,omitempty"`
	Emails   []string `json:"emails,omitempty"`
}

type SearchAccountsRequest struct {
	Query string
	OpenCursor bool
	CursorID string
	TimeoutMS int
}

type SearchAccountsResponse struct {
	GigyaResponse
	Results      []Account `json:"results"`
	ObjectsCount int       `json:"objectsCount"`
	TotalCount   int       `json:"totalCount"`
	NextCursorID string    `json:"nextCursorId,omitempty"`
}

func parseGigyaResponse(body []byte, successData any) error {
	var envelope GigyaResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("gigya: failed to parse response envelope: %w", err)
	}

	if envelope.IsError() {
		return &APIError{Response: envelope, Raw: body}
	}

	if successData != nil {
		if err := json.Unmarshal(body, successData); err != nil {
			return fmt.Errorf("gigya: failed to parse success body: %w", err)
		}
	}

	return nil
}
