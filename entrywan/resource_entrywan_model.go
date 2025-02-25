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

func modelResource() *schema.Resource {
	return &schema.Resource{
		Description:   "AI model.  More information at https://www.entrywan.com/docs#models",
		CreateContext: resourceModelCreate,
		ReadContext:   resourceModelRead,
		UpdateContext: resourceModelUpdate,
		DeleteContext: resourceModelDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A handy name for remembering which model is which.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"location": {
				Description: "The physical data center the model operates in.  us1 only during alpha.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "Model type.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"state": {
				Description: "Model state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"token": {
				Description: "Model token.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"endpoint": {
				Description: "Model endpoint.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

type modelCreateRes struct {
	Id string `json:"id"`
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	name := d.Get("name").(string)
	location := d.Get("location").(string)
	modelType := d.Get("type").(string)
	client := http.Client{}
	var jb []byte
	jb = []byte(fmt.Sprintf(`{"name": "%s", "location": "%s", "type": "%s"}`, name, location, modelType))
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/model", br)
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
		return diag.Errorf("unable to create model: %s", string(b))
	}
	var cr modelCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceModelRead(ctx, d, m)
}

type modelGetRes struct {
	State    string `json:"state"`
	Id       string `json:"id"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/model/"+id, nil)
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
	var cr modelGetRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	d.Set("state", cr.State)
	d.Set("endpoint", cr.Endpoint)
	d.Set("token", cr.Token)
	return nil
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceModelRead(ctx, d, m)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/model/%s", endpoint, id), nil)
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
