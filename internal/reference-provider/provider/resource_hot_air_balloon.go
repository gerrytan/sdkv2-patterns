package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHotAirBalloon() *schema.Resource {
	return &schema.Resource{
		Description:   "A hot air balloon flying in the sky.",
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

func resourceHotAirBalloonCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	result, err := meta.(*HelloConfig).Client.request(ctx, http.MethodPost, "hello_hot_air_balloon", "", hotAirBalloonPayload(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(result["id"].(string))
	return resourceHotAirBalloonRead(ctx, d, meta)
}

func resourceHotAirBalloonRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	result, err := meta.(*HelloConfig).Client.request(ctx, http.MethodGet, "hello_hot_air_balloon", d.Id(), nil)
	if mockNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	for _, key := range []string{"name", "sku", "additional_gas_tank", "exterior_colour", "modification"} {
		if err := d.Set(key, result[key]); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceHotAirBalloonUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if _, err := meta.(*HelloConfig).Client.request(ctx, http.MethodPut, "hello_hot_air_balloon", d.Id(), hotAirBalloonPayload(d)); err != nil {
		return diag.FromErr(err)
	}
	return resourceHotAirBalloonRead(ctx, d, meta)
}

func resourceHotAirBalloonDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if _, err := meta.(*HelloConfig).Client.request(ctx, http.MethodDelete, "hello_hot_air_balloon", d.Id(), nil); err != nil && !mockNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func hotAirBalloonPayload(d *schema.ResourceData) map[string]any {
	return map[string]any{
		"name":                d.Get("name"),
		"sku":                 d.Get("sku"),
		"additional_gas_tank": d.Get("additional_gas_tank"),
		"exterior_colour":     d.Get("exterior_colour").(*schema.Set).List(),
		"modification":        d.Get("modification").(*schema.Set).List(),
	}
}
