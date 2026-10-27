package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePaperPlane() *schema.Resource {
	return &schema.Resource{
		Description:   "A state-only paper plane for demonstrating SDKv2 patterns.",
		CreateContext: resourcePaperPlaneCreate,
		ReadContext:   resourcePaperPlaneRead,
		UpdateContext: resourcePaperPlaneUpdate,
		DeleteContext: resourcePaperPlaneDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the paper plane.",
			},
			"exterior_colour": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The exterior colour of the paper plane.",
			},
			"sku": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The SKU of the paper plane. Must be standard or premium. Default value 'standard' is set by remote API if not set",
				ValidateFunc: validation.StringInSlice([]string{"standard", "premium"}, false),
			},
		},
	}
}

func resourcePaperPlaneCreate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	if _, ok := d.GetOk("sku"); !ok {
		if err := d.Set("sku", "standard"); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourcePaperPlaneRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func resourcePaperPlaneUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId(d.Get("name").(string))
	return nil
}

func resourcePaperPlaneDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")
	return nil
}
