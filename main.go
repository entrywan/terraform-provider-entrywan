package main

import (
	"git.local/entrywan/terraform-provider-entrywan/entrywan"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: entrywan.Provider,
	})
}
