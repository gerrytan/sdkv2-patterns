package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// New returns the hello reference provider.
func New() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"hello_hot_air_balloon": resourceHotAirBalloon(),
		},
	}
}
