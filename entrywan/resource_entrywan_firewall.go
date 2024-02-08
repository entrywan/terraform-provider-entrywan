package entrywan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func firewallResource() *schema.Resource {
	return &schema.Resource{
		Description:   "Firewalls help secure compute instances by selectively allowing or denying certain kinds of traffic.  More information at https://entrywan.com/docs#firewall",
		CreateContext: resourceFirewallCreate,
		ReadContext:   resourceFirewallRead,
		UpdateContext: resourceFirewallUpdate,
		DeleteContext: resourceFirewallDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A handy name for remembering which firewall is which.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"rules": {
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Description: "Port number",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"src": {
							Description: "Source address of traffic",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"protocol": {
							Description: "Traffic protocol, either all, tcp, udp, icmp and a few others.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

type firewallCreateRes struct {
	Id string `json:"id"`
}

type Rule struct {
	Port     string
	Protocol string
	Src      string
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	name := d.Get("name").(string)
	rulesIface := d.Get("rules").([]interface{})
	rulesJson := []byte("[]")
	rulesJson, _ = json.Marshal(rulesIface)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(`{"name": "%s", "rules": %s}`, name, rulesJson))
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/firewall", br)
	if err != nil {
		fmt.Printf("error forming request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %v", err)
	}
	var b []byte
	b, err = ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return diag.Errorf("unable to create firewall: %s", string(b))
	}
	var cr firewallCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceFirewallRead(ctx, d, m)
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return nil
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceFirewallRead(ctx, d, m)
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/firewall/%s", endpoint, id), nil)
	if err != nil {
		fmt.Printf("error forming request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %v", err)
	}
	d.SetId("")
	return nil
}
