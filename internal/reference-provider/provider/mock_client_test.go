package provider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"sdkv2-patterns/internal/reference-provider/mockclient"

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
			client, err := mockclient.New()
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
			client, err = mockclient.New()
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
