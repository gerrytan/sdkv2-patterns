package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

func (c *mockClient) request(ctx context.Context, method, resourceType, id string, data map[string]any) (map[string]any, error) {
	var body bytes.Buffer
	if data != nil {
		if err := json.NewEncoder(&body).Encode(data); err != nil {
			return nil, err
		}
	}
	url := "http://hello.mock/" + resourceType
	if id != "" {
		url += "/" + id
	}
	req, err := http.NewRequestWithContext(ctx, method, url, &body)
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

var mockIDPattern = regexp.MustCompile(`^[0-9]{16}$`)

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/"), "/")
	if len(parts) > 2 {
		return mockResponse(req, http.StatusNotFound, map[string]any{"error": "invalid resource path"})
	}
	resourceType := parts[0]
	id := ""
	if len(parts) == 2 {
		id = parts[1]
		if !mockIDPattern.MatchString(id) {
			return mockResponse(req, http.StatusNotFound, map[string]any{"error": "resource not found"})
		}
	}
	if (req.Method == http.MethodPost && id != "") || (req.Method != http.MethodPost && id == "") {
		return mockResponse(req, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}

	mockDataMu.Lock()
	defer mockDataMu.Unlock()
	path := filepath.Join(t.directory, resourceType+"_"+id+".json")
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
	case http.MethodPost, http.MethodPut:
		var data map[string]any
		if req.Body == nil {
			return mockResponse(req, http.StatusBadRequest, map[string]any{"error": "expected a JSON object"})
		}
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil || data == nil {
			return mockResponse(req, http.StatusBadRequest, map[string]any{"error": "expected a JSON object"})
		}
		if resourceType == "hello_paper_plane" && (data["sku"] == nil || data["sku"] == "") {
			data["sku"] = "standard"
		}
		status := http.StatusOK
		if req.Method == http.MethodPost {
			if err := os.MkdirAll(t.directory, 0o755); err != nil {
				return nil, err
			}
			// Reserve the filename exclusively so even separate processes cannot
			// overwrite an existing resource if their random IDs collide.
			for {
				n, err := rand.Int(rand.Reader, big.NewInt(9_000_000_000_000_000))
				if err != nil {
					return nil, err
				}
				id = fmt.Sprintf("%d", n.Int64()+1_000_000_000_000_000)
				path = filepath.Join(t.directory, resourceType+"_"+id+".json")
				file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
				if errors.Is(err, os.ErrExist) {
					continue
				}
				if err != nil {
					return nil, err
				}
				if err := file.Close(); err != nil {
					return nil, err
				}
				break
			}
			status = http.StatusCreated
		} else if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return mockResponse(req, http.StatusNotFound, map[string]any{"error": "resource not found"})
		} else if err != nil {
			return nil, err
		}
		data["id"] = id
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
