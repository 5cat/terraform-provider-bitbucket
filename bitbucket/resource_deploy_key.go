package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDeployKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeployKeysCreate,
		Read:   resourceDeployKeysRead,
		Update: resourceDeployKeysUpdate,
		Delete: resourceDeployKeysDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceDeployKeysCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	deployKey := expandsshKey(d)
	log.Printf("[DEBUG] Deploy Key Request: %#v", deployKey)
	bytedata, err := json.Marshal(deployKey)

	if err != nil {
		return err
	}

	repo := d.Get("repository").(string)
	workspace := d.Get("workspace").(string)
	deployKeyReq, err := client.Post(fmt.Sprintf("2.0/repositories/%s/%s/deploy-keys", workspace, repo), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(deployKeyReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Deploy Keys Create Response JSON: %v", string(body))

	var deployKeyRes SshKey

	decodeerr := json.Unmarshal(body, &deployKeyRes)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Deploy Keys Create Response Decoded: %#v", deployKeyRes)

	d.SetId(string(fmt.Sprintf("%s/%s/%d", workspace, repo, deployKeyRes.ID)))

	return resourceDeployKeysRead(d, m)
}

func resourceDeployKeysRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	workspace, repo, keyId, err := deployKeyId(d.Id())
	if err != nil {
		return err
	}
	deployKeysReq, _ := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/deploy-keys/%s", workspace, repo, keyId))

	if deployKeysReq.StatusCode == 404 {
		log.Printf("[WARN] Deploy Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if deployKeysReq.Body == nil {
		return fmt.Errorf("error getting Deploy Key (%s): empty response", d.Id())
	}

	var deployKey *SshKey
	body, readerr := ioutil.ReadAll(deployKeysReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Deploy Key Response JSON: %v", string(body))

	decodeerr := json.Unmarshal(body, &deployKey)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Deploy Key Response Decoded: %#v", deployKey)

	d.Set("repository", repo)
	d.Set("workspace", workspace)
	d.Set("key", d.Get("key").(string))
	d.Set("label", deployKey.Label)
	d.Set("comment", deployKey.Comment)
	d.Set("key_id", deployKey.ID)

	return nil
}

func resourceDeployKeysUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	deployKey := expandsshKey(d)
	log.Printf("[DEBUG] Deploy Key Request: %#v", deployKey)
	bytedata, err := json.Marshal(deployKey)

	if err != nil {
		return err
	}

	workspace, repo, keyId, err := deployKeyId(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/deploy-keys/%s",
		workspace, repo, keyId), bytes.NewBuffer(bytedata))

	if err != nil {
		return fmt.Errorf("error updating Deploy Key (%s): %w", d.Id(), err)
	}

	return resourceDeployKeysRead(d, m)
}

func resourceDeployKeysDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	workspace, repo, keyId, err := deployKeyId(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(fmt.Sprintf("2.0/repositories/%s/%s/deploy-keys/%s", workspace, repo, keyId))

	if err != nil {
		return fmt.Errorf("error deleting Deploy Key (%s): %w", d.Id(), err)
	}

	return err
}

func deployKeyId(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected WORKSPACE-ID/REPO-ID/KEY-ID", id)
	}

	return parts[0], parts[1], parts[2], nil
}
