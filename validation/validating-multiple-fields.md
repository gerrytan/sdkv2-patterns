---
tocTitle: Validating multiple fields
tocSection: Validation
---

# Validating multiple fields

Use a `CustomizeDiff` function to validate multiple fields together. In the example below, the resource `hello_hot_air_balloon` only allows the `additional_gas_tank` field to be set when the `sku` is `premium`.

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon.go#schema" lang="go" -->
<!-- ref-end -->