package provider

import (
	"context"

	"sdkv2-patterns/internal/reference-provider/mockclient"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type HelloConfig struct {
	Client *mockclient.Client
}

func New() *schema.Provider {
	return &schema.Provider{
		ConfigureContextFunc: func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
			client, err := mockclient.New()
			return &HelloConfig{Client: client}, diag.FromErr(err)
		},
		ResourcesMap: map[string]*schema.Resource{
			"hello_hot_air_balloon": resourceHotAirBalloon(),
			"hello_paper_plane":     resourcePaperPlane(),
		},
	}
}
