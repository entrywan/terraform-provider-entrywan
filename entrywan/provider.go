package entrywan

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var token string
var endpoint string

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Description: "Entrywan IAM token",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ENTRYWAN_TOKEN", nil),
				Sensitive:   true,
			},
			"endpoint": {
				Description: "Entrywan API endpoint",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ENTRYWAN_ENDPOINT", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"entrywan_instance":     instanceResource(),
			"entrywan_sshkey":       sshkeyResource(),
			"entrywan_cluster":      clusterResource(),
			"entrywan_app":          appResource(),
			"entrywan_firewall":     firewallResource(),
			"entrywan_loadbalancer": loadbalancerResource(),
			"entrywan_vpc":          vpcResource(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	tokenVal := d.Get("token").(string)
	token = tokenVal

	endpointVal := d.Get("endpoint").(string)
	endpoint = endpointVal
	return nil, nil
}
