package avatar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrDisabled = errors.New("avatar provider is not configured")

type Response struct {
	URL string `json:"url"`
}

type Client struct {
	baseURL *url.URL
	http    *http.Client
}

func NewClient(rawURL string) (*Client, error) {
	if rawURL == "" {
		return nil, nil
	}
	baseURL, err := url.Parse(rawURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid avatar API URL")
	}
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 3 * time.Second}}, nil
}

func (c *Client) Get(ctx context.Context, seed string) (Response, error) {
	if c == nil {
		return Response{}, ErrDisabled
	}
	endpoint := *c.baseURL
	query := endpoint.Query()
	query.Set("seed", seed)
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Response{}, err
	}
	response, err := c.http.Do(request)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("avatar provider returned status %d", response.StatusCode)
	}

	var payload struct {
		Results []struct {
			Picture struct {
				Large string `json:"large"`
			} `json:"picture"`
		} `json:"results"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Response{}, err
	}
	if len(payload.Results) == 0 || payload.Results[0].Picture.Large == "" {
		return Response{}, errors.New("avatar provider returned no avatar")
	}
	return Response{URL: payload.Results[0].Picture.Large}, nil
}
