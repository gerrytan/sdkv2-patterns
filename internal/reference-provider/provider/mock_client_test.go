package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestMockResourceCRUD(t *testing.T) {
	for _, test := range []struct {
		resourceType string
		resource     *schema.Resource
		config       map[string]any
	}{
		{
			resourceType: "hello_hot_air_balloon",
			resource:     resourceHotAirBalloon(),
			config: map[string]any{
				"name":                "balloon",
				"sku":                 "premium",
				"additional_gas_tank": 3,
				"exterior_colour":     []any{"red", "blue"},
				"modification":        []any{map[string]any{"id": "satNav", "description": "navigation"}},
			},
		},
		{
			resourceType: "hello_paper_plane",
			resource:     resourcePaperPlane(),
			config:       map[string]any{"name": "plane", "exterior_colour": "white"},
		},
	} {
		t.Run(test.resourceType, func(t *testing.T) {
			t.Chdir(t.TempDir())
			ctx := context.Background()
			client, err := newMockClient()
			if err != nil {
				t.Fatal(err)
			}
			d := schema.TestResourceDataRaw(t, test.resource.Schema, test.config)
			config := &HelloConfig{Client: client}
			if diagnostics := test.resource.CreateContext(ctx, d, config); diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			id := d.Id()
			name := test.config["name"].(string)
			if id != test.resourceType+"/"+name {
				t.Fatalf("expected a type/name ID, got %q", id)
			}
			path := filepath.Join("hello_mock_client_data", test.resourceType+"_"+name+".json")
			readData := func() map[string]any {
				t.Helper()
				encoded, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var result map[string]any
				if err := json.Unmarshal(encoded, &result); err != nil {
					t.Fatal(err)
				}
				return result
			}
			stored := readData()
			if _, ok := stored["name"]; ok {
				t.Fatalf("name should only be in the URL, got persisted data: %#v", stored)
			}
			if _, ok := stored["id"]; ok {
				t.Fatalf("ID should not be generated in persisted data: %#v", stored)
			}
			if test.resourceType == "hello_paper_plane" && (stored["sku"] != "standard" || d.Get("sku") != "standard") {
				t.Fatal("API default SKU was not persisted and read into state")
			}
			// A fresh client and empty Terraform state must recover all fields.
			client, err = newMockClient()
			if err != nil {
				t.Fatal(err)
			}
			config = &HelloConfig{Client: client}
			fresh := schema.TestResourceDataRaw(t, test.resource.Schema, nil)
			fresh.SetId(id)
			if diagnostics := test.resource.ReadContext(ctx, fresh, config); diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			for key := range test.resource.Schema {
				original, restored := d.Get(key), fresh.Get(key)
				if set, ok := original.(*schema.Set); ok {
					original = set.List()
					restored = restored.(*schema.Set).List()
				}
				if !reflect.DeepEqual(original, restored) {
					t.Errorf("field %s did not survive persistence: %v != %v", key, d.Get(key), fresh.Get(key))
				}
			}
			if err := fresh.Set("sku", "premium"); err != nil {
				t.Fatal(err)
			}
			if diagnostics := test.resource.UpdateContext(ctx, fresh, config); diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			if fresh.Id() != id || readData()["sku"] != "premium" {
				t.Fatal("update did not preserve the ID and persist the SKU")
			}
			if diagnostics := test.resource.DeleteContext(ctx, fresh, config); diagnostics.HasError() {
				t.Fatal(diagnostics)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) || fresh.Id() != "" {
				t.Fatal("delete did not remove the file and clear the ID")
			}
			if diagnostics := test.resource.ReadContext(ctx, d, config); !diagnostics.HasError() || d.Id() != id {
				t.Fatalf("reading a missing resource should return an error and preserve the ID: %v", diagnostics)
			}
			fresh.SetId(id)
			if diagnostics := test.resource.DeleteContext(ctx, fresh, config); diagnostics.HasError() || fresh.Id() != "" {
				t.Fatalf("deleting a missing resource should succeed: %v", diagnostics)
			}
		})
	}
}

func TestMockClientPUTUpsert(t *testing.T) {
	t.Chdir(t.TempDir())
	client, err := newMockClient()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// Resource types do not need to be registered with the transport.
	for _, resourceType := range []string{"hello_future_resource", "hello_another_resource"} {
		for _, value := range []string{"created", "updated"} {
			if _, err := client.request(ctx, http.MethodPut, resourceType, "same-name", map[string]any{"value": value}); err != nil {
				t.Fatal(err)
			}
			result, err := client.request(ctx, http.MethodGet, resourceType, "same-name", nil)
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
	_, err = client.request(ctx, http.MethodPost, "hello_future_resource", "same-name", map[string]any{"value": "unexpected"})
	var apiError *mockHTTPError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected POST to return HTTP 405, got %v", err)
	}
}
