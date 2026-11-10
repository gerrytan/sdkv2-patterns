package provider

import (
	"context"
	"net/http"

	"sdkv2-patterns/internal/reference-provider/mockclient"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAeroplane() *schema.Resource {
	return &schema.Resource{
		Description:   "An aeroplane.",
		CreateContext: resourceAeroplaneCreate,
		ReadContext:   resourceAeroplaneRead,
		DeleteContext: resourceAeroplaneDelete,
		// ref-start: case-insensitive-api-schema
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the aeroplane.",
			},
			"registration_code": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The registration code of the aeroplane.",
			},
		},
		// ref-end: case-insensitive-api-schema
	}
}

func resourceAeroplaneCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	name := d.Get("name").(string)
	payload := map[string]any{
		"name":              name,
		"registration_code": d.Get("registration_code"),
	}
	if _, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, "hello_aeroplane", name, payload); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("hello_aeroplane/" + name)
	return resourceAeroplaneRead(ctx, d, meta)
}

func resourceAeroplaneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	return diag.FromErr(d.Set("registration_code", result["registration_code"].(string)))
}

func resourceAeroplaneDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
