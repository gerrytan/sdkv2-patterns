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

Following is an example on how to use a `CustomizeDiff` function to validate multiple fields together[^1]:

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#validating-multi-fields-customdiff" lang="go" -->

```go
func resourceHotAirBalloonAdditionalGasTankSkuCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	sku, tanks := "", 0

	if val, ok := d.GetOkExists("sku"); ok {
		sku = val.(string)
	} else {
		return nil
	}

	if val, ok := d.GetOkExists("additional_gas_tank"); ok {
		tanks = val.(int)
	} else {
		return nil
	}

	if sku != "premium" && tanks > 0 {
		return fmt.Errorf("additional_gas_tank can only be set when sku is premium")
	}
	return nil
}
```

<!-- ref-end -->

`GetOkExists` will check whether value is known in CustomizeDiff as well as having the default value applied. See [customized-diff-default-unknown-values.md](customized-diff-default-unknown-values.md) for more info.

[^1]: Refer to [`TestAccHotAirBalloonAdditionalGasTankInvalid`](../internal/reference-provider/provider/resource_hot_air_balloon_test.go).
