---
tocTitle: Using list for set
tocSection: Schema design
---

# Using list for set

Often representing set collection as list in Terraform is preferrable due to known issues with set in SDKv2 (refer to [working with set](working-with-set.md)).

<!-- ref-start ref="internal/reference-provider/provider/resource_paper_plane.go#using-list-for-set-schema" lang="go" -->

```go
"plugins": {
	Type:        schema.TypeList,
	Optional:    true,
	MinItems:    1,
	Description: "Plugins for the paper plane, API does not preserve order and always returns them sorted by ID.",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the plugin.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the plugin.",
			},
		},
	},
},
```

<!-- ref-end -->

## Uniqueness validation

If appropriate, implement a CustomizedDiff uniqueness validator at plan time[^1]

<!-- ref-start ref="internal/reference-provider/provider/resource_paper_plane.go#using-list-for-set-customdiff-dup-validator" lang="go" -->

```go
func resourcePaperPlanePluginsCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	plugins := config.GetAttr("plugins")
	if plugins.IsNull() || !plugins.IsKnown() {
		return nil
	}
	seen := make(map[string]bool)
	for iterator := plugins.ElementIterator(); iterator.Next(); {
		_, plugin := iterator.Element()
		if plugin.IsNull() || !plugin.IsKnown() {
			continue
		}
		id := plugin.GetAttr("id")
		if id.IsNull() || !id.IsKnown() {
			continue
		}
		if seen[id.AsString()] {
			return fmt.Errorf("plugins contains duplicate id %q", id.AsString())
		}
		seen[id.AsString()] = true
	}
	return nil
}
```

<!-- ref-end -->

## Ordering

Sometimes ordering is not preserved by API and Terraform will show plan diff right after creation. In the example below, 2 `plugins` are provided: `b` then `a`, but the API (erroneously) return them in alphabetical order: `a` then `b`.

<!-- ref-start file="internal/reference-provider/provider/testdata/resource_paper_plane_test_plugins_order.tf" lang="terraform" -->

```terraform
resource "hello_paper_plane" "test" {
  name = "test-paper-plane"

  plugins {
    id          = "b"
    description = "Plugin b"
  }

  plugins {
    id          = "a"
    description = "Plugin a"
  }
}
```

<!-- ref-end -->

Right after creation, a confusing plan diff wanting to reoder back into `b` then `a` will be shown[^2].

Implement a `DiffSuppressFunc` to ignore ordering differences.

<!-- ref-start ref="internal/reference-provider/provider/resource_paper_plane.go#using-list-for-set-diff-suppress" lang="go" -->

```go
func resourcePaperPlanePluginsV2DiffSuppress(_ string, _ string, _ string, d *schema.ResourceData) bool {
	oldValue, newValue := d.GetChange("plugins_v2")
	oldPlugins, newPlugins := oldValue.([]any), newValue.([]any)
	if len(oldPlugins) != len(newPlugins) {
		return false
	}
	type plugin struct {
		id          string
		description string
	}
	counts := make(map[plugin]int)
	for _, value := range oldPlugins {
		entry := value.(map[string]any)
		counts[plugin{entry["id"].(string), entry["description"].(string)}]++
	}
	for _, value := range newPlugins {
		entry := value.(map[string]any)
		key := plugin{entry["id"].(string), entry["description"].(string)}
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	return true
}
```

<!-- ref-end -->

## Value references

Value references of list element where order is not preserved by API can result in `Provider produced inconsistent final plan` errors as shown in the example below:

<!-- ref-start file="internal/reference-provider/provider/testdata/resource_paper_plane_test_plugins_order_value_ref.tf" lang="terraform" -->

```terraform
resource "hello_paper_plane" "test" {
  name = "test-paper-plane"

  plugins_v2 {
    id          = "b"
    description = "Plugin b"
  }

  plugins_v2 {
    id          = "a"
    description = "Plugin a"
  }
}

resource "hello_paper_plane" "test2" {
  name = "test-paper-plane"

  plugins_v2 {
    id          = hello_paper_plane.test.plugins_v2[0].id
    description = hello_paper_plane.test.plugins_v2[0].description
  }
}
```

<!-- ref-end -->

When both `test` and `test2` are deployed the first time, `hello_paper_plane.test.plugins_v2[0].id` is computed into `b`
at plan time, but became `a` after API response is received[^3]. This results in an error.

If the API ordering is deterministic, a better approach is to implement a `CustomizeDiff` validator that forces the config ordering.

<!-- code ref -->

[^1]: Refer to [`TestAccPaperPlanePluginsDuplicatesInvalid`](../internal/reference-provider/provider/resource_paper_plane_test.go)
[^2]: Refer to [`TestAccPaperPlanePluginsOrder`](../internal/reference-provider/provider/resource_paper_plane_test.go)
[^3]: Refer to [`TestAccPaperPlanePluginsV2Order`](../internal/reference-provider/provider/resource_paper_plane_test.go)