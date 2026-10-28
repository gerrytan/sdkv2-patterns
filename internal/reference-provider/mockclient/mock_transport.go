package mockclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type mockTransport struct {
	directory string
}

var mockDataMu sync.Mutex

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	resourceType, name, err := ParseResourceID(strings.TrimPrefix(req.URL.Path, "/"))
	if err != nil {
		return mockResponse(req, http.StatusNotFound, map[string]any{"error": "invalid resource path"})
	}

	mockDataMu.Lock()
	defer mockDataMu.Unlock()
	path := filepath.Join(t.directory, resourceType+"_"+name+".json")
	switch req.Method {
	case http.MethodGet:
		return t.readDataFromFile(req, path)
	case http.MethodPut:
		return t.storeDataIntoFile(req, path)
	case http.MethodDelete:
		return t.deleteDataFile(req, path)
	default:
		return mockResponse(req, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

func (t *mockTransport) readDataFromFile(req *http.Request, path string) (*http.Response, error) {
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
}

func (t *mockTransport) storeDataIntoFile(req *http.Request, path string) (*http.Response, error) {
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
}

func (t *mockTransport) deleteDataFile(req *http.Request, path string) (*http.Response, error) {
	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		return mockResponse(req, http.StatusNotFound, map[string]any{"error": "resource not found"})
	} else if err != nil {
		return nil, err
	}
	return mockResponse(req, http.StatusNoContent, nil)
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
