package provider

import (
	"context"
	"log"
	"net/http"

	"sdkv2-patterns/internal/reference-provider/mockclient"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKite() *schema.Resource {
	return &schema.Resource{
		Description:   "A kite flying in the sky.",
		CreateContext: resourceKiteCreate,
		ReadContext:   resourceKiteRead,
		UpdateContext: resourceKiteUpdate,
		DeleteContext: resourceKiteDelete,
		CustomizeDiff: resourceKiteCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the kite.",
			},
			"sku": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "standard",
				Description: "The SKU of the kite. Must be standard or premium. Defaults to standard.",
			},
		},
	}
}

func resourceKiteCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if sku, ok := d.GetOkExists("sku"); ok {
		log.Printf("hello_kite CustomizeDiff: name=%q sku=%q", d.Get("name").(string), sku.(string))
	} else {
		log.Printf("hello_kite CustomizeDiff: name=%q sku is not known", d.Get("name").(string))
	}

	return nil
}

func resourceKiteCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	name := d.Get("name").(string)
	_, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, "hello_kite", name, kitePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("hello_kite/" + name)
	return resourceKiteRead(ctx, d, meta)
}

func resourceKiteRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	resourceType, name, err := mockclient.ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	result, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodGet, resourceType, name, nil)
	if mockclient.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sku", result["sku"]); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceKiteUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	resourceType, name, err := mockclient.ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, resourceType, name, kitePayload(d)); err != nil {
		return diag.FromErr(err)
	}
	return resourceKiteRead(ctx, d, meta)
}

func resourceKiteDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	resourceType, name, err := mockclient.ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodDelete, resourceType, name, nil); err != nil && !mockclient.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func kitePayload(d *schema.ResourceData) map[string]any {
	return map[string]any{
		"sku": d.Get("sku"),
	}
}
