package gigya

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

type AccountsService struct {
	client *Client
}

func (s *AccountsService) Search(ctx context.Context, req SearchAccountsRequest) (*SearchAccountsResponse, error) {
	if req.Query == "" && req.CursorID == "" {
		return nil, fmt.Errorf("gigya: accounts.search requires Query or CursorID")
	}

	params := url.Values{}
	if req.Query != "" {
		params.Set("query", req.Query)
	}
	if req.OpenCursor {
		params.Set("openCursor", "true")
	}
	if req.CursorID != "" {
		params.Set("cursorId", req.CursorID)
	}
	if req.TimeoutMS > 0 {
		params.Set("timeout", strconv.Itoa(req.TimeoutMS))
	}

	resp, err := s.client.doRequest(ctx, "/accounts.search", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gigya: failed to read accounts.search response: %w", err)
	}

	var result SearchAccountsResponse
	if err := parseGigyaResponse(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *AccountsService) SearchAll(ctx context.Context, query string) ([]Account, error) {
	var all []Account

	resp, err := s.Search(ctx, SearchAccountsRequest{Query: query, OpenCursor: true})
	if err != nil {
		return nil, err
	}
	all = append(all, resp.Results...)

	cursorID := resp.NextCursorID
	for cursorID != "" {
		resp, err = s.Search(ctx, SearchAccountsRequest{CursorID: cursorID})
		if err != nil {
			return all, err
		}
		all = append(all, resp.Results...)
		cursorID = resp.NextCursorID
	}

	return all, nil
}
