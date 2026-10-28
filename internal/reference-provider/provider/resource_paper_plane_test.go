package provider

import (
	_ "embed"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//go:embed testdata/resource_paper_plane_test_default_sku.tf
var testAccPaperPlaneDefaultSKUConfig string

func TestAccPaperPlaneDefaultSKU(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccPaperPlaneDefaultSKUConfig,
			Check:  resource.TestCheckResourceAttr("hello_paper_plane.test", "sku", "standard"),
		}},
	})
}
