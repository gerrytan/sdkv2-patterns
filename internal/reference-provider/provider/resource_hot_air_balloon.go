package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHotAirBalloon() *schema.Resource {
	return &schema.Resource{
		Description:   "A state-only hot air balloon for demonstrating SDKv2 patterns.",
		CreateContext: resourceHotAirBalloonCreate,
		ReadContext:   resourceHotAirBalloonRead,
		UpdateContext: resourceHotAirBalloonUpdate,
		DeleteContext: resourceHotAirBalloonDelete,
		CustomizeDiff: customdiff.Sequence(
			resourceHotAirBalloonCustomizeDiff,
		),
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the hot air balloon.",
			},
			// ref-start: validating-multi-fields-schema
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
			// ref-end: validating-multi-fields-schema
			// ref-start: working-with-set-schema
			"exterior_colour": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The exterior colours of the hot air balloon.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			// ref-end: working-with-set-schema
			// ref-start: complex-set-schema
			"modification": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Modifications to the hot air balloon, identified by id.",
				Set: func(value any) int {
					return schema.HashString(value.(map[string]any)["id"])
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The identifier of the modification.",
						},
						"description": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The description of the modification.",
						},
					},
				},
			},
		},
		// ref-end: complex-set-schema
	}
}

// ref-start: validating-multi-fields-customdiff
func resourceHotAirBalloonCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	tanks := config.GetAttr("additional_gas_tank")
	if tanks.IsNull() || !tanks.IsKnown() {
		return nil
	}
	sku := config.GetAttr("sku")
	if sku.IsNull() || !sku.IsKnown() {
		return nil
	}
	if sku.AsString() != "premium" {
		return fmt.Errorf("additional_gas_tank can only be set when sku is premium")
	}
	return nil
}

// ref-end: validating-multi-fields-customdiff

func resourceHotAirBalloonCreate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceHotAirBalloonRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func resourceHotAirBalloonUpdate(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceHotAirBalloonDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")
	return nil
}
