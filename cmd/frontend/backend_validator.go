package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var backendBaseURL = "http://localhost:8080"

type PostStatusChecker interface {
	GetPostStatus(ctx context.Context, postID int64, cookieHeader string) (int, error)
}

var postStatusChecker PostStatusChecker = NewBackendStatusClient(backendBaseURL, 3*time.Second)

type BackendStatusClient struct {
	baseURL string
	client  *http.Client
}

func NewBackendStatusClient(baseURL string, timeout time.Duration) *BackendStatusClient {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	return &BackendStatusClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *BackendStatusClient) GetPostStatus(
	ctx context.Context,
	postID int64,
	cookieHeader string,
) (int, error) {
	return c.GetStatus(ctx, fmt.Sprintf("/api/v1/posts/%d", postID), cookieHeader)
}

func (c *BackendStatusClient) GetStatus(
	ctx context.Context,
	apiPath string,
	cookieHeader string,
) (int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+normalizeAPIPath(apiPath),
		nil,
	)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)

	return res.StatusCode, nil
}

func normalizeAPIPath(apiPath string) string {
	if strings.HasPrefix(apiPath, "/") {
		return apiPath
	}
	return "/" + apiPath
}
