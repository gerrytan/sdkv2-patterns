---
tocTitle: Default and unknown values in CustomizeDiff
tocSection: Validation
---

# Handling default and unknown values in CustomizeDiff

`CustomizeDiff` function should abstain if the value being validated is unknown. In the configuration example below, the `sku` value of `hello_kite.test` is not known during plan because the resource it depends on has not yet been created. There should not be any validation error thrown during plan time.

```hcl
resource "hello_plan" "test" {
  name = "test-plan"
  // outputs a "sku" attribute after creation
}

resource "hello_kite" "test" {
  name = "test-kite"
  sku  = hello_plan.test.sku
}
```

The schema for `sku` is a string with default value `"standard"`

```go
"sku": {
  Type:         schema.TypeString,
  Optional:     true,
  Default:      "standard",
  Description:  "The SKU of the kite. Defaults to standard.",
},
```

`d.Get("sku")` for `hello_kite.test` inside `CustomizeDiff` will return the go empty value for the string type which is empty string `""`.

```go
func resourceKiteCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
  sku := d.Get("sku").(string)
  if sku == "" {
    return nil
  }
  // further logic here...
}
```

`d.GetOk("sku")` behaves the same plus additional `ok` value returned which only indicates a non-zero value is set

```go
func resourceKiteCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
  if sku, ok := d.GetOk("sku"); ok {
    // "sku" is a non-zero value
  } else {
    // sku is a zero value, either unknown or explicitly set to ""
    return nil
  }

  // further logic here...
}
```

If user set a zero value for the type (eg: empty string `""` for string type), the `ok` value above will also be false, making it impossible to tell if the value is unknown, or explicitly set to the zero value.

```terraform
resource "hello_kite" "test" {
  name = "test-kite"
  sku  = ""
}
```

`d.GetOkExists("sku")` can be used to tell apart zero value from unknown value:

```go
func resourceKiteCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
  if sku, ok := d.GetOkExists("sku"); ok {
    // "sku" is a known value, which could be a zero value for the type (eg: "" for string)
  } else {
    // sku is unknown value
    return nil
  }

  // further logic here...
}
```

All get `d.GetOk..` functions has default value computation done. As shown in schema above, `"sku"` will default to `"standard"` if not specified set. The value `"standard"` will be returned in the CustomizeDiff for this configuration:

```terraform
resource "hello_kite" "test" {
  name = "test-kite"
}
```
