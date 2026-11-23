package provider

import (
  _ "embed"
  "testing"

  "github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//go:embed testdata/resource_aeroplane_test.tf
var testAccAeroplaneConfig string

func TestAccAeroplaneCaseSensitivity(t *testing.T) {
  resource.Test(t, resource.TestCase{
    ProviderFactories: testAccProviderFactories,
    Steps: []resource.TestStep{
      {
        Config: testAccAeroplaneConfig,
      },
      {
        Config: `
resource "hello_aeroplane" "test" {
  name              = "test-aeroplane"
  registration_code = "abc123"
}
`,
        PlanOnly:           true,
        ExpectNonEmptyPlan: true,
        ConfigPlanChecks: resource.ConfigPlanChecks{
          PostApplyPreRefresh: []plancheck.PlanCheck{
            plancheck.ExpectResourceAction("hello_aeroplane.test", plancheck.ResourceActionReplace),
          },
        },
      },
    },
  })
}

func TestAccAeroplaneEngineTypeOnly(t *testing.T) {
  resource.Test(t, resource.TestCase{
    ProviderFactories: testAccProviderFactories,
    Steps: []resource.TestStep{{
      Config: `
resource "hello_aeroplane" "test" {
  name              = "test-aeroplane-engine-type-only"
  registration_code = "ABC123"

  engine {
    type = "propeller"
  }
}
`,
      Check: resource.ComposeTestCheckFunc(
        resource.TestCheckResourceAttr("hello_aeroplane.test", "engine.0.type", "propeller"),
        resource.TestCheckResourceAttr("hello_aeroplane.test", "engine.0.speed", "low"),
      ),
      ExpectNonEmptyPlan: true,
    }},
  })
}
