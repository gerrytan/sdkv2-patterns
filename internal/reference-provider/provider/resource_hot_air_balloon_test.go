package provider

import (
	_ "embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"hello": func() (*schema.Provider, error) { return New(), nil },
}

func testAccBalloonConfig(sku, tanks string) string {
	return fmt.Sprintf(`
resource "hello_hot_air_balloon" "test" {
  name = "test-balloon"
  sku  = %q
  %s
}
`, sku, tanks)
}

func TestAccHotAirBalloonAdditionalGasTank(t *testing.T) {
	for _, test := range []struct {
		name  string
		sku   string
		tanks string
	}{
		{name: "standard", sku: "standard"},
		{name: "premium", sku: "premium", tanks: "additional_gas_tank = 1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: testAccProviderFactories,
				Steps: []resource.TestStep{{
					Config: testAccBalloonConfig(test.sku, test.tanks),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("hello_hot_air_balloon.test", "name", "test-balloon"),
						resource.TestCheckResourceAttr("hello_hot_air_balloon.test", "sku", test.sku),
					),
				}},
			})
		})
	}
}

func TestAccHotAirBalloonAdditionalGasTankInvalid(t *testing.T) {
	for _, test := range []struct {
		name  string
		sku   string
		tanks int
		error string
	}{
		{name: "standard_with_tanks", sku: "standard", tanks: 1, error: "additional_gas_tank can only be set when sku is premium"},
		{name: "below_minimum", sku: "premium", tanks: 0, error: "expected additional_gas_tank to be in the range"},
		{name: "negative", sku: "premium", tanks: -1, error: "expected additional_gas_tank to be in the range"},
		{name: "above_maximum", sku: "premium", tanks: 11, error: "expected additional_gas_tank to be in the range"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: testAccProviderFactories,
				Steps: []resource.TestStep{{
					Config:      testAccBalloonConfig(test.sku, fmt.Sprintf("additional_gas_tank = %d", test.tanks)),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(test.error),
				}},
			})
		})
	}
}

//go:embed resource_hot_air_balloon_test_value_unknown.tf
var testAccHotAirBalloonValueUnknownConfig string

func TestAccHotAirBalloonInvalidAdditionalGasTankAllowedIfValueIsUnknown(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccHotAirBalloonValueUnknownConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:      testAccHotAirBalloonValueUnknownConfig,
				ExpectError: regexp.MustCompile("additional_gas_tank can only be set when sku is premium"),
			},
		},
	})
}
