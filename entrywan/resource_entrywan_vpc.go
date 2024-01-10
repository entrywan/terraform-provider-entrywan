package entrywan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func vpcResource() *schema.Resource {
	return &schema.Resource{
		Description: "Virtual Private Cloud for establishing encrypted private networks for instances.  More information at https://entrywan.com/docs#vpcnetworks",
		Create:      resourceVpcCreate,
		Read:        resourceVpcRead,
		Update:      resourceVpcUpdate,
		Delete:      resourceVpcDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A handy name for remembering which VPC is which.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"prefix": {
				Description: "The CIDR prefix of the network.  Example: 192.168.5.0/24",
				Required:    true,
				Type:        schema.TypeString,
			},
			"members": {
				Description: "The initial members of the VPC.",
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip4public": {
							Description: "The public IPv4 address of the instance.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"ip4private": {
							Description: "The private IPv4 address of the instance.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
					},
				},
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
	for _, memberIface := range d.Get("members").([]any) {
		member := memberIface.(map[string]any)
		ip4public := member["ip4public"].(string)
		ip4private := member["ip4private"].(string)
		var jb []byte
		if ip4private == "" {
			jb = []byte(fmt.Sprintf(`{"ip4public": "%s"}`, ip4public))
		} else {
			jb = []byte(fmt.Sprintf(`{"ip4public": "%s", "ip4private": "%s"}`, ip4public, ip4private))
		}
		br := bytes.NewReader(jb)
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/vpc/%s", endpoint, cr.Id), br)
		if err != nil {
			fmt.Printf("error forming request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		_, err = client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v", err)
		}
	}
	return resourceVpcRead(d, m)
}

func resourceVpcRead(d *schema.ResourceData, m any) error {
	name := d.Get("name")
	if err := d.Set("name", name); err != nil {
		return err
	}
	return nil
}

type vpcmember struct {
	Ippublic  string `json: "ippublic"`
	Ipprivate string `json: "ipprivate"`
}

type vpcRes struct {
	Id      string      `json: "id"`
	Members []vpcmember `json: "members"`
}

func resourceVpcUpdate(d *schema.ResourceData, m any) error {
	if d.HasChange("members") {
		client := http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/vpc", endpoint), nil)
		if err != nil {
			fmt.Printf("error forming request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %v", err)
		}
		b, err := io.ReadAll(res.Body)
		var vpcs []vpcRes
		err = json.Unmarshal(b, &vpcs)
		id := d.Id()
		targetMembers := d.Get("members").([]any)
		for _, targetMemberIface := range targetMembers {
			targetMember := targetMemberIface.(map[string]any)
			found := false
			for _, vpc := range vpcs {
				if vpc.Id != id {
					continue
				}
				for _, vpcMember := range vpc.Members {
					if vpcMember.Ippublic == targetMember["ip4public"] {
						found = true
					}
				}
			}
			if !found {
				ip4public := targetMember["ip4public"].(string)
				ip4private := targetMember["ip4private"].(string)
				var jb []byte
				if ip4private == "" {
					jb = []byte(fmt.Sprintf(`{"ip4public": "%s"}`, ip4public))
				} else {
					jb = []byte(fmt.Sprintf(`{"ip4public": "%s", "ip4private": "%s"}`, ip4public, ip4private))
				}
				br := bytes.NewReader(jb)
				req, err := http.NewRequest("PUT", fmt.Sprintf("%s/vpc/%s", endpoint, id), br)
				if err != nil {
					fmt.Printf("error forming request: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
				_, err = client.Do(req)
				if err != nil {
					fmt.Printf("error making request: %v", err)
				}
			}
		}
		for _, vpc := range vpcs {
			if vpc.Id != id {
				continue
			}
			for _, vpcMember := range vpc.Members {
				found := false
				for _, targetMemberIface := range targetMembers {
					targetMember := targetMemberIface.(map[string]any)
					if vpcMember.Ippublic == targetMember["ip4public"] {
						found = true
					}
				}
				if !found {
					jb := []byte(fmt.Sprintf(`{"ip4private": "%s"}`, vpcMember.Ipprivate))
					br := bytes.NewReader(jb)
					req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/vpc/%s", endpoint, id), br)
					if err != nil {
						fmt.Printf("error forming request: %v", err)
					}
					req.Header.Set("Authorization", "Bearer "+token)
					_, err = client.Do(req)
					if err != nil {
						fmt.Printf("error making request: %v", err)
					}
				}
			}
		}
	}
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
