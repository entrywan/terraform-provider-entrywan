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

func clusterResource() *schema.Resource {
	return &schema.Resource{
		Description:   "Kubernetes cluster comprised of control plane and worker nodes.  More information at https://entrywan.com/docs#kubernetes",
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A handy name for remembering which cluster is which.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location": {
				Description: "The physical data center the cluster operates in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"size": {
				Description: "The number of worker nodes.  Can be scaled up or down as needed.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"state": {
				Description: "Cluster state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"apiserver": {
				Description: "Cluster API server IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"version": {
				Description: "Cluster version.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cni": {
				Description: "The networking plugin to use, either flannel or calico.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

type clusterCreateRes struct {
	Id string `json:"id"`
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
	if res.StatusCode != 200 {
		return diag.Errorf("unable to create cluster: %s", string(b))
	}
	var cr clusterCreateRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	return resourceClusterRead(ctx, d, m)
}

type clusterGetRes struct {
	State     string `json:"state"`
	Apiserver string `json:"apiserver"`
	Version   string `json:"version"`
	Id        string `json:"id"`
	Size      int    `json:"size"`
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	id := d.Id()
	client := http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/cluster/"+id, nil)
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
	var cr clusterGetRes
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Printf("error unmarshaling request: %v", err)
	}
	d.SetId(cr.Id)
	d.Set("state", cr.State)
	d.Set("apiserver", cr.Apiserver)
	d.Set("version", cr.Version)
	d.Set("size", cr.Size)
	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	if d.HasChange("size") {
		size := d.Get("size")
		id := d.Id()
		client := http.Client{}
		jb := []byte(fmt.Sprintf(`{"size": %d}`, size))
		br := bytes.NewReader(jb)
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/cluster/%s/scale", endpoint, id), br)
		if err != nil {
			fmt.Printf("error forming request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		_, err = client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v", err)
		}
	}
	return resourceClusterRead(ctx, d, m)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
