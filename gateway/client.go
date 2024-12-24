package gateway

import (
	"context"
	"net/http"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(tr http.RoundTripper) *HTTPClient {
	client := &http.Client{
		Transport: tr,
	}
	return &HTTPClient{client}
}

func (h *HTTPClient) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err = h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}
