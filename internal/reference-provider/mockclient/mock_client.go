package mockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	httpClient *http.Client
}

func New() (*Client, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &Client{httpClient: &http.Client{
		Transport: &mockTransport{directory: filepath.Join(pwd, "hello_mock_client_data")},
	}}, nil
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("mock API returned HTTP %d: %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	var apiError *HTTPError
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound
}

func (c *Client) Request(ctx context.Context, method, resourceType, name string, data map[string]any) (map[string]any, error) {
	var body bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&body).Encode(data); err != nil {
			return nil, err
		}
	}
	endpoint := "http://hello.mock/" + url.PathEscape(resourceType) + "/" + url.PathEscape(name)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		message, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, &HTTPError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(message))}
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func ParseResourceID(id string) (resourceType, name string, err error) {
	resourceType, name, ok := strings.Cut(id, "/")
	if !ok || resourceType == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid resource ID %q: expected <resource_type>/<name>", id)
	}
	return resourceType, name, nil
}
