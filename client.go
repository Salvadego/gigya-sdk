package gigya

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultDataCenter = "us1"
	defaultTimeout    = 60 * time.Second
	apiVersion        = "2"
)

type AuthConfig struct {
	APIKey string

	ClientID     string
	ClientSecret string

	UserKey string
	Secret  string
}

func (a AuthConfig) usesOAuth2() bool {
	return a.ClientID != "" && a.ClientSecret != ""
}

type ClientConfig struct {
	DataCenter string

	BaseURL string

	HTTPClient *http.Client
	UserAgent  string
}

type Client struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
	authConfig AuthConfig

	tokenMu sync.RWMutex
	token   string

	Auth     *AuthService
	Accounts *AccountsService
}

func NewClient(authConfig AuthConfig, clientConfig *ClientConfig) *Client {
	if clientConfig == nil {
		clientConfig = &ClientConfig{}
	}

	dc := clientConfig.DataCenter
	if dc == "" {
		dc = defaultDataCenter
	}

	baseURL := clientConfig.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://accounts.%s.gigya.com", dc)
	}

	httpClient := clientConfig.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	userAgent := clientConfig.UserAgent
	if userAgent == "" {
		userAgent = "gigya-go-client/1.0"
	}

	client := &Client{
		baseURL:    baseURL,
		userAgent:  userAgent,
		httpClient: httpClient,
		authConfig: authConfig,
	}

	client.Auth = &AuthService{client: client}
	client.Accounts = &AccountsService{client: client}

	return client
}

func (c *Client) Token() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

func (c *Client) SetToken(token string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = token
}

func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	if params == nil {
		params = url.Values{}
	}

	params.Set("format", "json")
	params.Set("apiKey", c.authConfig.APIKey)

	useBearer := c.authConfig.usesOAuth2()

	reqURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if useBearer {
		params.Set("httpStatusCodes", "false")
	} else {
		if err := signParams(reqURL, params, c.authConfig.UserKey, c.authConfig.Secret); err != nil {
			return nil, fmt.Errorf("gigya: failed to sign request: %w", err)
		}
	}

	body := strings.NewReader(params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if useBearer {
		token := c.Token()
		if token == "" {
			return nil, fmt.Errorf("%w: call client.Auth.Authenticate(ctx) first", ErrNotAuthenticated)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}

	return resp, nil
}

func (c *Client) rawDoRequest(ctx context.Context, endpoint string, body io.Reader, headers map[string]string) (*http.Response, error) {
	reqURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}

	return resp, nil
}
