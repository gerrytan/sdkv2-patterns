---
tocTitle: List vs set
tocSection: Schema Design
---

# List vs set

Set implicitly enforces uniqueness. Duplicate values can still be specified, but they will be automatically deduplicated. Consider a set field `exterior_colour` in `hello_hot_air_balloon`:

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#list-vs-set-schema" lang="go" -->

```go
"exterior_colour": {
	Type:        schema.TypeSet,
	Optional:    true,
	Description: "The exterior colours of the hot air balloon.",
	Elem:        &schema.Schema{Type: schema.TypeString},
},
```

<!-- ref-end -->

Although 3 `exterior_colour`values are specified, only 2 values are stored into state because "red" is duplicated[^1]:

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

[^1] Refer to [`TestAccHotAirBalloonExteriorColourDuplicates`](../internal/reference-provider/provider/resource_hot_air_balloon_test.go)