package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func loadbalancerResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadbalancerCreate,
		Read:   resourceLoadbalancerRead,
		Update: resourceLoadbalancerUpdate,
		Delete: resourceLoadbalancerDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"algo": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
			},
			"listeners": {
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"targets": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"ip": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type loadbalancerCreateRes struct {
	Id string `json:"id"`
}

func resourceLoadbalancerCreate(d *schema.ResourceData, m any) error {
	name := d.Get("name").(string)
	location := d.Get("location").(string)
	algo := d.Get("algo").(string)
	protocol := d.Get("protocol").(string)
	listenersIface := d.Get("listeners").([]interface{})
	listenersJson := []byte("[]")
	listenersJson, _ = json.Marshal(listenersIface)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(`{"name": "%s", "location": "%s", "algo": "%s", "protocol": "%s", "listeners": %s}`,
		name,
		location,
		algo,
		protocol,
		listenersJson))
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/loadbalancer", br)
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
	var cr loadbalancerCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceLoadbalancerRead(d, m)
}

func resourceLoadbalancerRead(d *schema.ResourceData, m any) error {
	name := d.Get("name")
	if err := d.Set("name", name); err != nil {
		return err
	}
	return nil
}

func resourceLoadbalancerUpdate(d *schema.ResourceData, m any) error {
	id := d.Id()
	if d.HasChange("listeners") {
		listenersIface := d.Get("listeners").([]interface{})
		listenersJson := []byte("[]")
		listenersJson, _ = json.Marshal(listenersIface)
		client := http.Client{}
		jb := []byte(fmt.Sprintf(`{"listeners": %s}`,
			listenersJson))
		br := bytes.NewReader(jb)
		req, err := http.NewRequest("PUT", endpoint+"/loadbalancer/"+id, br)
		if err != nil {
			fmt.Printf("error forming request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		_, err = client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v", err)
		}
	}
	return resourceLoadbalancerRead(d, m)
}

func resourceLoadbalancerDelete(d *schema.ResourceData, m any) error {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/loadbalancer/%s", endpoint, id), nil)
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
