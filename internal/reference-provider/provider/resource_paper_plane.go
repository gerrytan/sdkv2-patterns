package provider

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"sdkv2-patterns/internal/reference-provider/mockclient"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		CustomizeDiff: customdiff.Sequence(
			resourcePaperPlanePluginsCustomizeDiff,
		),
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
			// ref-start: using-list-for-set-schema
			"plugins": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				Description: "Plugins for the paper plane, API does not preserve order and always returns them sorted by ID.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The identifier of the plugin.",
						},
						"description": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The description of the plugin.",
						},
					},
				},
			},
			// ref-end: using-list-for-set-schema
			"plugins_v2": {
				Type:             schema.TypeList,
				Optional:         true,
				MinItems:         1,
				Description:      "Same like plugins but with DiffSuppressFunc to ignore ordering differences",
				DiffSuppressFunc: resourcePaperPlanePluginsV2DiffSuppress,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The identifier of the plugin.",
						},
						"description": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The description of the plugin.",
						},
					},
				},
			},
		},
	}
}

// ref-start: using-list-for-set-customdiff-dup-validator
func resourcePaperPlanePluginsCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	plugins := config.GetAttr("plugins")
	if plugins.IsNull() || !plugins.IsKnown() {
		return nil
	}
	seen := make(map[string]bool)
	for iterator := plugins.ElementIterator(); iterator.Next(); {
		_, plugin := iterator.Element()
		if plugin.IsNull() || !plugin.IsKnown() {
			continue
		}
		id := plugin.GetAttr("id")
		if id.IsNull() || !id.IsKnown() {
			continue
		}
		if seen[id.AsString()] {
			return fmt.Errorf("plugins contains duplicate id %q", id.AsString())
		}
		seen[id.AsString()] = true
	}
	return nil
}

// ref-end: using-list-for-set-customdiff-dup-validator

// ref-start: using-list-for-set-diff-suppress
func resourcePaperPlanePluginsV2DiffSuppress(_ string, _ string, _ string, d *schema.ResourceData) bool {
	oldValue, newValue := d.GetChange("plugins_v2")
	oldPlugins, newPlugins := oldValue.([]any), newValue.([]any)
	if len(oldPlugins) != len(newPlugins) {
		return false
	}
	type plugin struct {
		id          string
		description string
	}
	counts := make(map[plugin]int)
	for _, value := range oldPlugins {
		entry := value.(map[string]any)
		counts[plugin{entry["id"].(string), entry["description"].(string)}]++
	}
	for _, value := range newPlugins {
		entry := value.(map[string]any)
		key := plugin{entry["id"].(string), entry["description"].(string)}
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	return true
}
// ref-end: using-list-for-set-diff-suppress

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
	// Simulate an API that always returns plugins sorted by ID, regardless
	// of the order submitted in the request payload.
	for _, key := range []string{"plugins", "plugins_v2"} {
		if plugins, ok := result[key].([]any); ok {
			sort.SliceStable(plugins, func(i, j int) bool {
				return plugins[i].(map[string]any)["id"].(string) < plugins[j].(map[string]any)["id"].(string)
			})
		}
	}
	for _, key := range []string{"sku", "exterior_colour", "plugins", "plugins_v2"} {
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
		"plugins":         d.Get("plugins"),
		"plugins_v2":      d.Get("plugins_v2"),
	}
}
