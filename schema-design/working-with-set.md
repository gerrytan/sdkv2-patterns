---
tocTitle: Working with set
tocSection: Schema Design
---

# Working with set

Set implicitly enforces uniqueness. Duplicate values can still be specified, but they will be automatically deduplicated. Consider a string set field `exterior_colour` in `hello_hot_air_balloon`:

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#working-with-set-schema" lang="go" -->

```go
"exterior_colour": {
	Type:        schema.TypeSet,
	Optional:    true,
	Description: "The exterior colours of the hot air balloon.",
	Elem:        &schema.Schema{Type: schema.TypeString},
},
```

<!-- ref-end -->

Although 3 `exterior_colour` values are specified in config, only 2 are stored into state because "red" is duplicated[^1]:

<!-- ref-start file="internal/reference-provider/provider/resource_hot_air_balloon_test_set_dup.tf" lang="terraform" -->

```terraform
resource "hello_hot_air_balloon" "test" {
  name            = "test-balloon"
  sku             = "standard"
  exterior_colour = ["red", "blue", "red"]
}
```

<!-- ref-end -->

If the values are replaced with `["blue", "red"]`, the plan diff will be empty because the set already contains these values[^1]. Also observe how the order of the elements is not significant.

## Element cannot be indexed

Set elements cannot be indexed and referenced by another resource. The example configuration below:

<!-- ref-start file="internal/reference-provider/provider/resource_hot_air_balloon_test_set_index.tf" lang="terraform" -->

```terraform
resource "hello_hot_air_balloon" "test" {
  name            = "test-balloon"
  sku             = "standard"
  exterior_colour = ["red", "blue"]
}

resource "hello_paper_plane" "test" {
  name            = "test-paper-plane"
  exterior_colour = hello_hot_air_balloon.test.exterior_colour[1]
}
```

<!-- ref-end -->

Will result in `Elements of a set are identified only by their value` plan time error[^2].

## Complex set with custom hash

The situation gets more confusing with a complex set with a custom hash. Consider a set field `modification` where only the `id` is used as element identifier:

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#complex-set-schema" lang="go" -->

```go
	"modification": {
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Modifications to the hot air balloon, identified by id.",
		Set: func(value any) int {
			return schema.HashString(value.(map[string]any)["id"])
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The identifier of the modification.",
				},
				"description": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The description of the modification.",
				},
			},
		},
	},
},
```

<!-- ref-end -->

If the user mistakenly supplied the same id twice, the last one wins and overwrites the previous one[^3].

<!-- ref-start file="internal/reference-provider/provider/resource_hot_air_balloon_test_modification_dup.tf" lang="terraform" -->

```terraform
resource "hello_hot_air_balloon" "test" {
  name = "test-balloon"
  sku  = "standard"

  modification { # this element will be overwritten by the one underneath because it has the same id!
    id          = "satNav"
    description = "satellite navigation"
  }

  modification {
    id          = "satNav"
    description = "yet another sat nav modification, mistakenly added"
  }
}
```

<!-- ref-end -->

[^1]: Refer to [`TestAccHotAirBalloonExteriorColourDuplicates`](../internal/reference-provider/provider/resource_hot_air_balloon_test.go)
[^2]: Refer to [`TestAccHotAirBalloonExteriorColourIndexInvalid`](../internal/reference-provider/provider/resource_hot_air_balloon_test.go)
[^3]: Refer to [`TestAccHotAirBalloonModificationDuplicates`](../internal/reference-provider/provider/resource_hot_air_balloon_test.go)