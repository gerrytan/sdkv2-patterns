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
    UpdateContext: resourceAeroplaneUpdate,
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
      // ref-end: case-insensitive-api-schema
      "engine": {
        Type:        schema.TypeList,
        MaxItems:    1,
        Optional:    true,
        Computed:    true,
        Description: "The engine of the aeroplane. API defaults to type propeller and speed low when omitted.",
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "type": {
              Type:        schema.TypeString,
              Required:    true,
              Description: "The type of the engine.",
            },
            "speed": {
              Type:        schema.TypeString,
              Optional:    true,
              Description: "The speed of the engine. API defaults to low when omitted.",
            },
          },
        },
      },
    },
  }
}

func resourceAeroplaneCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
  name := d.Get("name").(string)
  if _, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, "hello_aeroplane", name, aeroplanePayload(d)); err != nil {
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
  if err := d.Set("registration_code", result["registration_code"].(string)); err != nil {
    return diag.FromErr(err)
  }
  // Simulate the aeroplane API's server-side engine defaults.
  engine, ok := result["engine"].(map[string]any)
  if !ok {
    engine = map[string]any{"type": "propeller"}
  }
  if _, ok := engine["speed"]; !ok {
    engine["speed"] = "low"
  }
  return diag.FromErr(d.Set("engine", []any{engine}))
}

func resourceAeroplaneUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
  resourceType, name, err := mockclient.ParseResourceID(d.Id())
  if err != nil {
    return diag.FromErr(err)
  }
  if _, err := meta.(*HelloConfig).Client.Request(ctx, http.MethodPut, resourceType, name, aeroplanePayload(d)); err != nil {
    return diag.FromErr(err)
  }
  return resourceAeroplaneRead(ctx, d, meta)
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

func aeroplanePayload(d *schema.ResourceData) map[string]any {
  payload := map[string]any{
    "name":              d.Get("name"),
    "registration_code": d.Get("registration_code"),
  }
  if engines := d.Get("engine").([]any); len(engines) > 0 && engines[0] != nil {
    engine := engines[0].(map[string]any)
    value := map[string]any{"type": engine["type"]}
    if speed, ok := engine["speed"].(string); ok && speed != "" {
      value["speed"] = speed
    }
    payload["engine"] = value
  }
  return payload
}
