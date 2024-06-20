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

func instanceResource() *schema.Resource {
	return &schema.Resource{
		Description:   "A compute instance.  More information at https://www.entrywan.com/docs#instances",
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description: "The instance's hostname.  The machine is booted with this hostname on first boot.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location": {
				Description: "The physical data center the instance operates in.  Choose us1, us2 or uk1.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"disk": {
				Description: "Hard disk disk in GB.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"cpus": {
				Description: "Number of CPU cores.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"ram": {
				Description: "Memory in GB.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"sshkey": {
				Description: "The ssh key to be placed as authorized_keys on the machine.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"os": {
				Description: "The operating system image.  Choose alma, debian, fedora, rocky or ubuntu.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"userdata": {
				Description: "Optional script to run on first boot.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"state": {
				Description: "Instance state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ip4": {
				Description: "Instance primary IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vpcids": {
				Description: "Optional VPCs to attach the instance to.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type instanceCreateRes struct {
	Id  string `json:"id"`
	Ip4 string `json:"ip4"`
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	hostname := d.Get("hostname").(string)
	location := d.Get("location").(string)
	disk := d.Get("disk").(int)
	cpus := d.Get("cpus").(int)
	ram := d.Get("ram").(int)
	os := d.Get("os").(string)
	sshkey := d.Get("sshkey").(string)
	userdata := d.Get("userdata").(string)
	vpcIdsInt := d.Get("vpcids").([]interface{})
	vpcIds := make([]string, len(vpcIdsInt))
	for i, vpcIdInt := range vpcIdsInt {
		vpcIds[i] = vpcIdInt.(string)
	}
	var vpcIdsJson []byte
	vpcIdsJson, _ = json.Marshal(vpcIds)
	client := http.Client{}
	var jb []byte
	if len(vpcIds) > 0 {
		jb = []byte(fmt.Sprintf(
			`{"hostname": "%s",
         "vpcids": %s,
	 "location": "%s",
	 "disk": %d,
	 "cpus": %d,
	 "ram": %d,
	 "os": "%s",
	 "sshkeyname": "%s",
	 "userdata": %q}`,
			hostname, string(vpcIdsJson), location, disk, cpus, ram, os, sshkey, userdata))
	} else {
		jb = []byte(fmt.Sprintf(
			`{"hostname": "%s",
	 "location": "%s",
	 "disk": %d,
	 "cpus": %d,
	 "ram": %d,
	 "os": "%s",
	 "sshkeyname": "%s",
	 "userdata": %q}`,
			hostname, location, disk, cpus, ram, os, sshkey, userdata))
	}
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/instance", br)
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
		return diag.Errorf("unable to create instance: %s", string(b))
	}
	var cr instanceCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}

	d.SetId(cr.Id)
	d.Set("ip4", cr.Ip4)
	return resourceInstanceRead(ctx, d, m)
}

type instanceGetRes struct {
	State string `json:"state"`
	Id    string `json:"id"`
	Ip4   string `json:"ip4"`
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	ip4 := d.Get("ip4").(string)
	client := http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/instance/"+id, nil)
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
	var cr instanceGetRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	d.Set("state", cr.State)
	d.Set("ip4", ip4)
	return nil
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/instance/%s", endpoint, id), nil)
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
