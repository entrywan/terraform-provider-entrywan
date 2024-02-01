package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func instanceResource() *schema.Resource {
	return &schema.Resource{
		Description: "A compute instance.  More information at https://entrywan.com/docs#instances",
		Create:      resourceInstanceCreate,
		Read:        resourceInstanceRead,
		Update:      resourceInstanceUpdate,
		Delete:      resourceInstanceDelete,
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description: "The instance's hostname.  The machine is booted with this hostname on first boot.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location": {
				Description: "The physical data center the instance operates in.",
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
				Description: "The operating system image.",
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
		},
	}
}

type instanceCreateRes struct {
	Id  string `json:"id"`
	Ip4 string `json:"ip4"`
}

func resourceInstanceCreate(d *schema.ResourceData, m any) error {
	hostname := d.Get("hostname").(string)
	location := d.Get("location").(string)
	disk := d.Get("disk").(int)
	cpus := d.Get("cpus").(int)
	ram := d.Get("ram").(int)
	os := d.Get("os").(string)
	sshkey := d.Get("sshkey").(string)
	userdata := d.Get("userdata").(string)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(
		`{"hostname": "%s",
 "location": "%s",
 "disk": %d,
 "cpus": %d,
 "ram": %d,
 "os": "%s",
 "sshkeyname": "%s",
 "userdata": %q}`,
		hostname, location, disk, cpus, ram, os, sshkey, userdata))
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
	var cr instanceCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	d.Set("ip4", cr.Ip4)
	return resourceInstanceRead(d, m)
}

type instanceGetRes struct {
	State string `json:"state"`
	Id    string `json:"id"`
	Ip4   string `json:"ip4"`
}

func resourceInstanceRead(d *schema.ResourceData, m any) error {
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

func resourceInstanceUpdate(d *schema.ResourceData, m any) error {
	return resourceInstanceRead(d, m)
}

func resourceInstanceDelete(d *schema.ResourceData, m any) error {
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
