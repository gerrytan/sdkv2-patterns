package provider

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//go:embed testdata/resource_paper_plane_test_default_sku.tf
var testAccPaperPlaneDefaultSKUConfig string

//go:embed testdata/resource_paper_plane_test_plugins_dup.tf
var testAccPaperPlanePluginsDupConfig string

func TestAccPaperPlaneDefaultSKU(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccPaperPlaneDefaultSKUConfig,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("hello_paper_plane.test", "sku", "standard"),
				resource.TestCheckResourceAttr("hello_paper_plane.test", "id", "hello_paper_plane/test-paper-plane"),
			),
		}},
	})
}

func TestAccPaperPlanePluginsDuplicatesInvalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testAccPaperPlanePluginsDupConfig,
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`plugins contains duplicate id "satnav"`),
		}},
	})
}

//go:embed testdata/resource_paper_plane_test_plugins_order.tf
var testAccPaperPlanePluginsOrderConfig string

func TestAccPaperPlanePluginsOrder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{{
			Config:             testAccPaperPlanePluginsOrderConfig,
			Check:              resource.TestCheckResourceAttr("hello_paper_plane.test", "plugins.0.id", "a"),
			ExpectNonEmptyPlan: true,
		}},
	})
}

//go:embed testdata/resource_paper_plane_test_plugins_order_value_ref.tf
var testAccPaperPlanePluginsOrderValueRefConfig string

func TestAccPaperPlanePluginsV2Order(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPaperPlanePluginsOrderValueRefConfig,
				ExpectError: regexp.MustCompile(`(?s)Provider produced inconsistent final plan.*was\s+cty\.StringVal\("Plugin b"\),\s+but now\s+cty\.StringVal\("Plugin a"\)`),
			},
		},
	})
}

//go:embed testdata/resource_paper_plane_test_stickers.tf
var testAccPaperPlaneStickersConfig string

func TestAccPaperPlaneStickersTopLevelOnlyForceNew(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPaperPlaneStickersConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.#", "2"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.0.id", "sunrise"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.0.description", "A golden sunrise"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.1.id", "stars"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.1.description", "A sky full of stars"),
				),
			},
			{
				Config: strings.ReplaceAll(testAccPaperPlaneStickersConfig, "A golden sunrise", "A colourful sunrise"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hello_paper_plane.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.#", "2"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.0.id", "sunrise"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.0.description", "A colourful sunrise"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.1.id", "stars"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.1.description", "A sky full of stars"),
				),
			},
			{
				Config: `
resource "hello_paper_plane" "test" {
  name = "test-paper-plane-stickers"

  stickers {
    id          = "sunrise"
    description = "A colourful sunrise"
  }
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hello_paper_plane.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.#", "1"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.0.id", "sunrise"),
					resource.TestCheckResourceAttr("hello_paper_plane.test", "stickers.0.description", "A colourful sunrise"),
				),
			},
		},
	})
}
