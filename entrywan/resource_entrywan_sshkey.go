package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func sshkeyResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceSshkeyCreate,
		Read:   resourceSshkeyRead,
		Update: resourceSshkeyUpdate,
		Delete: resourceSshkeyDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pub": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

type sshkeyCreateRes struct {
	Id string `json:"id"`
}

func resourceSshkeyCreate(d *schema.ResourceData, m any) error {
	name := d.Get("name").(string)
	pub := d.Get("pub").(string)
	client := http.Client{}
	jb := []byte(fmt.Sprintf(`{"name": "%s", "pub": "%s"}`, name, pub))
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/sshkey", br)
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
	var cr sshkeyCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling response: %v", err)
	}
	d.SetId(cr.Id)
	return resourceSshkeyRead(d, m)
}

func resourceSshkeyRead(d *schema.ResourceData, m any) error {
	pub := d.Get("pub")

	if err := d.Set("pub", pub); err != nil {
		return err
	}
	return nil
}

func resourceSshkeyUpdate(d *schema.ResourceData, m any) error {
	return resourceSshkeyRead(d, m)
}

func resourceSshkeyDelete(d *schema.ResourceData, m any) error {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/sshkey/%s", endpoint, id), nil)
	if err != nil {
		fmt.Print("error forming request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %v", err)
	}
	return nil
}
