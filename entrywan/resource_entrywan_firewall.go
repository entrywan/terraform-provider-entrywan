package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func firewallResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallCreate,
		Read:   resourceFirewallRead,
		Update: resourceFirewallUpdate,
		Delete: resourceFirewallDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rules": {
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"src": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
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

func resourceFirewallCreate(d *schema.ResourceData, m any) error {
	name := d.Get("name").(string)
	rulesIface := d.Get("rules").([]interface{})
	rulesJson := []byte("[]")
	rulesJson, _ = json.Marshal(rulesIface)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(`{"name": "%s", "rules": %s}`, name, rulesJson))
	fmt.Println(string(jb))
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
	var cr firewallCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceFirewallRead(d, m)
}

func resourceFirewallRead(d *schema.ResourceData, m any) error {
	name := d.Get("name")
	if err := d.Set("name", name); err != nil {
		return err
	}
	return nil
}

func resourceFirewallUpdate(d *schema.ResourceData, m any) error {
	return resourceFirewallRead(d, m)
}

func resourceFirewallDelete(d *schema.ResourceData, m any) error {
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
