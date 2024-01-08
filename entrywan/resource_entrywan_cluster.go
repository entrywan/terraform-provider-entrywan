package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func clusterResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"cni": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

type clusterCreateRes struct {
	Id string `json:"id"`
}

func resourceClusterCreate(d *schema.ResourceData, m any) error {
	name := d.Get("name").(string)
	location := d.Get("location").(string)
	size := d.Get("size").(int)
	cni := d.Get("cni").(string)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(
		`{"name": "%s",
 "location": "%s",
 "size": %d,
 "cni": "%s"}`,
		name, location, size, cni))
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/cluster", br)
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
	var cr clusterCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceClusterRead(d, m)
}

func resourceClusterRead(d *schema.ResourceData, m any) error {
	name := d.Get("name")
	if err := d.Set("name", name); err != nil {
		return err
	}
	return nil
}

func resourceClusterUpdate(d *schema.ResourceData, m any) error {
	return resourceClusterRead(d, m)
}

func resourceClusterDelete(d *schema.ResourceData, m any) error {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/cluster/%s", endpoint, id), nil)
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
