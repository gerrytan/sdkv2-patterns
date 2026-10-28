package provider

import (
	"context"
	"net/http"

	"sdkv2-patterns/internal/reference-provider/mockclient"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePaperPlane() *schema.Resource {
	return &schema.Resource{
		Description:   "You fold it, you throw it.",
		CreateContext: resourcePaperPlaneCreate,
		ReadContext:   resourcePaperPlaneRead,
		UpdateContext: resourcePaperPlaneUpdate,
		DeleteContext: resourcePaperPlaneDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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

func resourcePaperPlaneCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	name := d.Get("name").(string)
	_, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, "hello_paper_plane", name, paperPlanePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("hello_paper_plane/" + name)
	return resourcePaperPlaneRead(ctx, d, meta)
}

func resourcePaperPlaneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	resourceType, name, err := mockclient.ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	result, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodGet, resourceType, name, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}
	for _, key := range []string{"sku", "exterior_colour"} {
		if err := d.Set(key, result[key]); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourcePaperPlaneUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	resourceType, name, err := mockclient.ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, resourceType, name, paperPlanePayload(d)); err != nil {
		return diag.FromErr(err)
	}
	return resourcePaperPlaneRead(ctx, d, meta)
}

func resourcePaperPlaneDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func paperPlanePayload(d *schema.ResourceData) map[string]any {
	sku := d.Get("sku").(string)
	if sku == "" {
		sku = "standard"
	}
	return map[string]any{
		"sku":             sku,
		"exterior_colour": d.Get("exterior_colour"),
	}
}
