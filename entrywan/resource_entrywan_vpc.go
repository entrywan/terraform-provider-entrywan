package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func vpcResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceVpcCreate,
		Read:   resourceVpcRead,
		Update: resourceVpcUpdate,
		Delete: resourceVpcDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"prefix": {
				Required: true,
				Type:     schema.TypeString,
			},
		},
	}
}

type vpcCreateRes struct {
	Id string `json:"id"`
}

func resourceVpcCreate(d *schema.ResourceData, m any) error {
	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(`{"name": "%s", "prefix": "%s"}`, name, prefix))
	fmt.Println(string(jb))
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/vpc", br)
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
	var cr vpcCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceVpcRead(d, m)
}

func resourceVpcRead(d *schema.ResourceData, m any) error {
	name := d.Get("name")
	if err := d.Set("name", name); err != nil {
		return err
	}
	return nil
}

func resourceVpcUpdate(d *schema.ResourceData, m any) error {
	return resourceVpcRead(d, m)
}

func resourceVpcDelete(d *schema.ResourceData, m any) error {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/vpc/%s", endpoint, id), nil)
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
