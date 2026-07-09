package gigya

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type AuthService struct {
	client *Client
}

func (s *AuthService) Authenticate(ctx context.Context) (TokenResponse, error) {
	if !s.client.authConfig.usesOAuth2() {
		return TokenResponse{}, fmt.Errorf("gigya: Authenticate requires AuthConfig.ClientID and ClientSecret")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	basicAuth := base64.StdEncoding.EncodeToString(
		[]byte(s.client.authConfig.ClientID + ":" + s.client.authConfig.ClientSecret),
	)

	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + basicAuth,
	}

	resp, err := s.client.rawDoRequest(ctx, "/oauth2/token", strings.NewReader(form.Encode()), headers)
	if err != nil {
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("gigya: failed to read token response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return TokenResponse{}, fmt.Errorf("gigya: oauth2/token http %s: %s", resp.Status, string(body))
	}

	var tokenResp TokenResponse
	if err := parseGigyaResponse(body, &tokenResp); err != nil {
		if uerr := jsonUnmarshalToken(body, &tokenResp); uerr == nil && tokenResp.AccessToken != "" {
			s.client.SetToken(tokenResp.AccessToken)
			return tokenResp, nil
		}
		return TokenResponse{}, err
	}

	s.client.SetToken(tokenResp.AccessToken)
	return tokenResp, nil
}
