---
tocTitle: Validating multiple fields
tocSection: Validation
---

# Validating multiple fields

Use a `CustomizeDiff` function to validate multiple fields together. In the example below, the resource `hello_hot_air_balloon` only allows the `additional_gas_tank` field to be set when the `sku` is `premium`.

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#validating-multi-fields-schema" lang="go" -->

```go
"sku": {
	Type:         schema.TypeString,
	Required:     true,
	Description:  "The SKU of the hot air balloon. Must be standard or premium.",
	ValidateFunc: validation.StringInSlice([]string{"standard", "premium"}, false),
},
"additional_gas_tank": {
	Type:         schema.TypeInt,
	Optional:     true,
	Description:  "Additional gas tanks, from 1 to 10. Only available with the premium SKU.",
	ValidateFunc: validation.IntBetween(1, 10),
},
```

<!-- ref-end -->

Following is an example on how to use a `CustomizeDiff` function to validate multiple fields together:

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#validating-multi-fields-customdiff" lang="go" -->

```go
func resourceHotAirBalloonCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	tanks := config.GetAttr("additional_gas_tank")
	if tanks.IsNull() || !tanks.IsKnown() {
		return nil
	}
	sku := config.GetAttr("sku")
	if sku.IsNull() || !sku.IsKnown() {
		return nil
	}
	if sku.AsString() != "premium" {
		return fmt.Errorf("additional_gas_tank can only be set when sku is premium")
	}
	return nil
}
```

<!-- ref-end -->

Observe how `IsKnown` value check is performed on each fields, and the validator opts out by returning nil early if the value is unknown. Sometimes values are not known during plan time and the validation logic cannot be done yet until apply time.

In the below example, `hello_hot_air_balloon` will pass validation during plan time, but the `can only be set when sku is premium` validation error will occur during apply time.

<!-- ref-start file="internal/reference-provider/provider/resource_hot_air_balloon_test_value_unknown.tf" lang="terraform" -->
<!-- ref-end -->