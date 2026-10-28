package mockclient

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
)

func TestMockClientPUTUpsert(t *testing.T) {
	t.Chdir(t.TempDir())
	client, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, resourceType := range []string{"hello_future_resource", "hello_another_resource"} {
		for _, value := range []string{"created", "updated"} {
			if _, err := client.Request(ctx, http.MethodPut, resourceType, "same-name", map[string]any{"value": value}); err != nil {
				t.Fatal(err)
			}
			result, err := client.Request(ctx, http.MethodGet, resourceType, "same-name", nil)
			if err != nil {
				t.Fatal(err)
			}
			if result["value"] != value {
				t.Fatalf("expected %q, got %#v", value, result)
			}
		}
	}
	entries, err := os.ReadDir("hello_mock_client_data")
	if err != nil || len(entries) != 2 {
		t.Fatalf("expected one file per resource type and name, got %v: %v", entries, err)
	}
	_, err = client.Request(ctx, http.MethodPost, "hello_future_resource", "same-name", map[string]any{"value": "unexpected"})
	var apiError *HTTPError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected POST to return HTTP 405, got %v", err)
	}
}
