package provider

import (
	_ "embed"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
