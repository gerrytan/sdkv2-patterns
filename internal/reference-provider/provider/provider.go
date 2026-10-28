package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type HelloConfig struct {
	Client *mockClient
}

func New() *schema.Provider {
	return &schema.Provider{
		ConfigureContextFunc: func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
			client, err := newMockClient()
			return &HelloConfig{Client: client}, diag.FromErr(err)
		},
		ResourcesMap: map[string]*schema.Resource{
			"hello_hot_air_balloon": resourceHotAirBalloon(),
			"hello_paper_plane":     resourcePaperPlane(),
		},
	}
}
