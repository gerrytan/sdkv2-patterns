package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func New() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"hello_hot_air_balloon": resourceHotAirBalloon(),
			"hello_paper_plane":     resourcePaperPlane(),
		},
	}
}
