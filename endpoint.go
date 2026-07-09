package gigya

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

type Endpoint struct {
	client    *Client
	namespace string
}

func (c *Client) Endpoint(namespace string) *Endpoint {
	return &Endpoint{client: c, namespace: namespace}
}

func (e *Endpoint) Sub(sub string) *Endpoint {
	return &Endpoint{client: e.client, namespace: e.namespace + "." + sub}
}

func (e *Endpoint) Call(ctx context.Context, method string, params url.Values, out any) error {
	call := e.namespace + "." + method

	resp, err := e.client.doRequest(ctx, "/"+call, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gigya: failed to read %s response: %w", call, err)
	}

	return parseGigyaResponse(body, out)
}

func (c *Client) AccountsEndpoint() *Endpoint {
	return c.Endpoint("accounts")
}
