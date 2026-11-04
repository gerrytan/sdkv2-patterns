---
tocTitle: Test assertions
tocSection: Testing
---

# Test assertions

One common mistake when implementing a resource is forgetting to set a property in the create / update API payload, or setting the state for a property when reading API response.

A bug is simulated in the `hello_hot_air_balloon` resource for the `exterior_colour` property: if the colour "rainbow" is set, it will be dropped from the create / update payload.

A plain acc test step alone will normally catch the problem without any checks because it requires an empty plan after every apply.

If it doesn't (for example the step expects nonempty plan), an `ImportStateVerify: true` step can be added underneath. This step works by importing the resource created in the previous step, and checking all the state received matches the configuration.

<!-- ref-start ref="internal/reference-provider/provider/resource_hot_air_balloon_test.go#test-assertions-import-state-verify" lang="go" -->

```go
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// An intentional bug was introduced where the colour "rainbow" is dropped from the API payload
				Config: `
resource "hello_hot_air_balloon" "test" {
  name            = "test-balloon"
  sku             = "standard"
  exterior_colour = ["rainbow"]
}
`,
				// Setting this to false would have detected the problem, but intentionally set to true to test ImportStateVerify
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      "hello_hot_air_balloon.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
```

<!-- ref-end -->