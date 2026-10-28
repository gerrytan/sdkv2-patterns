package provider

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
	"sync"
)

type mockClient struct {
	httpClient *http.Client
}

func newMockClient() (*mockClient, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &mockClient{httpClient: &http.Client{
		Transport: &mockTransport{directory: filepath.Join(pwd, "hello_mock_client_data")},
	}}, nil
}

type mockHTTPError struct {
	StatusCode int
	Message    string
}

func (e *mockHTTPError) Error() string {
	return fmt.Sprintf("mock API returned HTTP %d: %s", e.StatusCode, e.Message)
}

func mockNotFound(err error) bool {
	var apiError *mockHTTPError
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound
}

func (c *mockClient) request(ctx context.Context, method, resourceType, name string, data map[string]any) (map[string]any, error) {
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
		return nil, &mockHTTPError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(message))}
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

type mockTransport struct {
	directory string
}

var mockDataMu sync.Mutex

func parseResourceID(id string) (resourceType, name string, err error) {
	resourceType, name, ok := strings.Cut(id, "/")
	if !ok || resourceType == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid resource ID %q: expected <resource_type>/<name>", id)
	}
	return resourceType, name, nil
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	resourceType, name, err := parseResourceID(strings.TrimPrefix(req.URL.Path, "/"))
	if err != nil {
		return mockResponse(req, http.StatusNotFound, map[string]any{"error": "invalid resource path"})
	}

	mockDataMu.Lock()
	defer mockDataMu.Unlock()
	path := filepath.Join(t.directory, resourceType+"_"+name+".json")
	switch req.Method {
	case http.MethodGet:
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			return mockResponse(req, http.StatusNotFound, map[string]any{"error": "resource not found"})
		}
		if err != nil {
			return nil, err
		}
		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return mockResponse(req, http.StatusOK, result)
	case http.MethodPut:
		var data map[string]any
		if req.Body == nil {
			return mockResponse(req, http.StatusBadRequest, map[string]any{"error": "expected a JSON object"})
		}
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil || data == nil {
			return mockResponse(req, http.StatusBadRequest, map[string]any{"error": "expected a JSON object"})
		}
		status := http.StatusOK
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			status = http.StatusCreated
		} else if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(t.directory, 0o755); err != nil {
			return nil, err
		}
		encoded, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
			return nil, err
		}
		return mockResponse(req, status, data)
	case http.MethodDelete:
		if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
			return mockResponse(req, http.StatusNotFound, map[string]any{"error": "resource not found"})
		} else if err != nil {
			return nil, err
		}
		return mockResponse(req, http.StatusNoContent, nil)
	default:
		return mockResponse(req, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func mockResponse(req *http.Request, status int, data map[string]any) (*http.Response, error) {
	var body bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&body).Encode(data); err != nil {
			return nil, err
		}
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(&body),
		Request:    req,
	}, nil
}
