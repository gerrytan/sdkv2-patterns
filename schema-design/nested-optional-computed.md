---
tocTitle: Nested optional computed
tocSection: Schema design
---

# Nested Optional + Computed property

Optional + Computed defined at the top level does not always propagate to child properties. Taking example of this schema:

```go
"engine": {
  Type:        schema.TypeList,
  MaxItems:    1,
  Optional:    true,
  Computed:    true,
  Description: "The engine of the aeroplane. Defaults to type propeller.",
  Elem: &schema.Resource{
    Schema: map[string]*schema.Schema{
      "type": {
        Type:        schema.TypeString,
        Required:    true,
        Description: "The type of the engine.",
      },
    },
  },
},
```

Suppose the API will default to the following payload if none is provided. Observe that "speed" subproperty is not yet supported:

```json
{
  "engine": {
    "type": "propeller",
    "speed": "low"
  }
}
```

Omitting "engine" in Terraform config will result in no diff as expected:

```terraform
resource "hello_aeroplane" "test" {
  name              = "test-aeroplane"
  registration_code = "abc123"
}
```

Now let's add support for `engine.speed` in Terraform just with `Optional: true` with no default:

```diff
"engine": {
	Type:        schema.TypeList,
	MaxItems:    1,
	Optional:    true,
	Computed:    true,
	Description: "The engine of the aeroplane. Defaults to type propeller and speed low when omitted.",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of the engine.",
			},
+			"speed": {
+				Type:        schema.TypeString,
+				Optional:    true,
+				Description: "The speed of the engine. Defaults to low when omitted.",
+			},
		},
	},
},
```

While config that entirely omits `engine` stays unaffected (no plan diff), config that specified only `engine.type` without `engine.speed` will see following diff:

```
~ engine {
    - speed = "low" -> null
      # (1 unchanged attribute hidden)
  }
```

In conclusion: Optional + Computed does not automatically propagate to nested child properties, it has to be explicitly defined[^1].

[^1]: Refer to [TestAccAeroplaneEngineTypeOnly](../internal/reference-provider/provider/resource_aeroplane_test.go)