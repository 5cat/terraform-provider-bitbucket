package bitbucket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type UserGroupMembership struct {
	UUID string `json:"uuid,omitempty"`
}

func resourceGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupMembershipsPut,
		Read:   resourceGroupMembershipsRead,
		Delete: resourceGroupMembershipsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group_slug": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGroupMembershipsPut(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient

	workspace := d.Get("workspace").(string)
	groupSlug := d.Get("group_slug").(string)
	uuid := d.Get("uuid").(string)

	_, err := client.PutOnly(fmt.Sprintf("1.0/groups/%s/%s/members/%s",
		workspace, groupSlug, uuid))
	if err != nil {
		return err
	}

	d.SetId(string(fmt.Sprintf("%s/%s/%s", workspace, groupSlug, uuid)))

	return resourceGroupMembershipsRead(d, m)
}

func resourceGroupMembershipsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient

	workspace, slug, uuid, err := groupMemberId(d.Id())
	if err != nil {
		return err
	}

	groupsReq, _ := client.Get(fmt.Sprintf("1.0/groups/%s/%s/members", workspace, slug))

	if groupsReq.StatusCode == 404 {
		log.Printf("[WARN] Group Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if groupsReq.Body == nil {
		return fmt.Errorf("error reading Group Membership (%s): empty response", d.Id())
	}

	var members []*UserGroupMembership

	body, readerr := ioutil.ReadAll(groupsReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Group Membership Response JSON: %v", string(body))

	decodeerr := json.Unmarshal(body, &members)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Group Membership Response Decoded: %#v", members)

	if len(members) == 0 {
		return fmt.Errorf("error getting Group Members (%s): empty response", d.Id())
	}

	var member *UserGroupMembership
	for _, mbr := range members {
		if mbr.UUID == uuid {
			member = mbr
			break
		}
	}

	if member == nil {
		return fmt.Errorf("error getting Group Member (%s): not found", d.Id())
	}

	log.Printf("[DEBUG] Group Member Response Decoded: %#v", member)

	d.Set("workspace", workspace)
	d.Set("group_slug", slug)
	d.Set("uuid", member.UUID)

	return nil
}

func resourceGroupMembershipsDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(Clients).httpClient

	workspace, slug, uuid, err := groupMemberId(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(fmt.Sprintf("1.0/groups/%s/%s/members/%s",
		workspace, slug, uuid))

	if err != nil {
		return err
	}

	return err
}

func groupMemberId(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected WORKSPACE-ID/GROUP-SLUG-ID/MEMBER-UUID", id)
	}

	return parts[0], parts[1], parts[2], nil
}
