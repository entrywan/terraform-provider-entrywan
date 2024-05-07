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

func appResource() *schema.Resource {
	return &schema.Resource{
		Description:   "PaaS application.  Either repo- or OCI-based.  More information at https://www.entrywan.com/docs#apps",
		CreateContext: resourceAppCreate,
		ReadContext:   resourceAppRead,
		UpdateContext: resourceAppUpdate,
		DeleteContext: resourceAppDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The subdomain the app listens on, example: myapp.entrywan.app.  Must be globally unique.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"location": {
				Description: "The physical data center the app operates in.  us1 only during beta.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"size": {
				Description: "Amount of RAM in MB.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"state": {
				Description: "App state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				Description: "Port your app listens for HTTP traffic on.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"source": {
				Description: "Type of app to deploy, either github or oci.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"repo": {
				Description: "Required for repo-based apps, the repository URL.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"repobranch": {
				Description: "Required for repo-based apps, the repo branch name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"reporoot": {
				Description: "For repo-based apps, the optional directory root the app source begins at.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"credential": {
				Description: "For repo-based apps hosted in private repositories, a personal access token that grants at least read privileges to that repo.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"image": {
				Description: "Required for OCI-based apps, the image repository location.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

type appCreateRes struct {
	Id string `json:"id"`
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	name := d.Get("name").(string)
	location := d.Get("location").(string)
	size := d.Get("size").(int)
	port := d.Get("port").(int)
	source := d.Get("source").(string)
	image := d.Get("image").(string)
	repo := d.Get("repo").(string)
	repobranch := d.Get("repobranch").(string)
	reporoot := d.Get("reporoot").(string)
	credential := d.Get("credential").(string)
	client := http.Client{}
	var jb []byte
	if source == "oci" {
		jb = []byte(fmt.Sprintf(
			`{"name": "%s",
 "location": "%s",
 "image": "%s",
 "size": %d,
 "port": %d,
 "source": "%s"}`,
			name, location, image, size, port, source))
	} else {
		if credential == "" {
			jb = []byte(fmt.Sprintf(
				`{"name": "%s",
 "location": "%s",
 "repo": "%s",
 "repobranch": "%s",
 "reporoot": "%s",
 "size": %d,
 "port": %d,
 "source": "%s"}`,
				name, location, repo, repobranch, reporoot, size, port, source))
		} else {
			jb = []byte(fmt.Sprintf(
				`{"name": "%s",
 "location": "%s",
 "repo": "%s",
 "repobranch": "%s",
 "reporoot": "%s",
 "credential": "%s",
 "size": %d,
 "port": %d,
 "source": "%s"}`,
				name, location, repo, repobranch, reporoot, credential, size, port, source))
		}
	}
	br := bytes.NewReader(jb)
	req, err := http.NewRequest("POST", endpoint+"/app", br)
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
		return diag.Errorf("unable to create app: %s", string(b))
	}
	var cr appCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceAppRead(ctx, d, m)
}

type appGetRes struct {
	State string `json:"state"`
	Id    string `json:"id"`
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/app/"+id, nil)
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
	var cr appGetRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	d.Set("state", cr.State)
	return nil
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if d.HasChange("image") {
		image := d.Get("image")
		id := d.Id()
		client := http.Client{}
		jb := []byte(fmt.Sprintf(`{"image": "%s"}`, image))
		br := bytes.NewReader(jb)
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/app/%s", endpoint, id), br)
		if err != nil {
			fmt.Printf("error forming request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		_, err = client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v", err)
		}
	}
	return resourceAppRead(ctx, d, m)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/app/%s", endpoint, id), nil)
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
