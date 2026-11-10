---
tocTitle: Case insensitive API
tocSection: Schema Design
---

# Case insensitive API

Some API behave in case-insensitive way. In the "aeroplane" resource below, the registration code is case-insensitive, "ABC123" means the same thing as "abc123"

<!-- ref-start ref="internal/reference-provider/provider/resource_aeroplane.go#case-insensitive-api-schema" lang="go" -->

```go
Schema: map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The name of the aeroplane.",
	},
	"registration_code": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The registration code of the aeroplane.",
	},
},
```

<!-- ref-end -->

Consider a first deployment of registration code "ABC123":

<!-- ref-start file="internal/reference-provider/provider/testdata/resource_aeroplane_test.tf" lang="terraform" -->

```terraform
resource "hello_aeroplane" "test" {
  name              = "test-aeroplane"
  registration_code = "ABC123"
}
```

<!-- ref-end -->

After this is deployed, the value "ABC123" is stored into state.

Subsequent deployment with registration code "abc123" might be required (for example: replacing hardcoded value with data source that always returns lower case). This will result in catastrophic destroy and re-create.

If implementing new resource, the proper solution is to normalize the value in the read function before setting into state. In the example below the "normal" value is all lowercase:

```go
func resourceAeroplaneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// ...
	return diag.FromErr(d.Set("registration_code", strings.ToLower(regCodeFromAPI)))
}
```

But this cannot be done when fixing bug against an existing resource. Any config that use different casing than normal will see a confusing plan diff after updating the provider version.

Instead we have to implement a `DiffSuppressFunc` that ignores case differences:

```go
"registration_code": {
	Type:        schema.TypeString,
	Required:    true,
	ForceNew:    true,
	Description: "The registration code of the aeroplane.",
	DiffSuppressFunc: func(_ string, old, new string, _ *schema.ResourceData) bool {
		return strings.EqualFold(old, new)
	},
},
```