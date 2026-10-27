package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"sdkv2-patterns/internal/reference-provider/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New,
	})
}
