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
		Description: "Layer 3 load balancer for distributing network traffic among healthy instances.  More information at https://entrywan.com/docs#loadbalancers",
		Create:      resourceLoadbalancerCreate,
		Read:        resourceLoadbalancerRead,
		Update:      resourceLoadbalancerUpdate,
		Delete:      resourceLoadbalancerDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A handy name for remembering which load balancer is which.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"location": {
				Description: "The physical data center the load balancer operates in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"algo": {
				Description: "Load balancing algorithm to choose, either round-robin or least-used.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"protocol": {
				Description: "Traffic protocol, either tcp or http.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ip": {
				Description: "Load balancer primary IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"listeners": {
				Description: "A listener for each port the load balancer should respond to traffic on.",
				Required:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Description: "Port number.",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"targets": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Description: "Target port number.",
										Type:        schema.TypeInt,
										Required:    true,
									},
									"ip": {
										Description: "Target IP address or hostname",
										Type:        schema.TypeString,
										Required:    true,
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

type loadbalancerGetRes struct {
	Id string `json:"id"`
	Ip string `json:"ip"`
}

func resourceLoadbalancerRead(d *schema.ResourceData, m any) error {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/loadbalancer/"+id, nil)
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
	var cr loadbalancerGetRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	d.Set("ip", cr.Ip)
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
