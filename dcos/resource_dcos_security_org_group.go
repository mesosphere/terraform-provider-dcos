package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosSecurityOrgGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosSecurityOrgGroupCreate,
		Read:   resourceDcosSecurityOrgGroupRead,
		Update: resourceDcosSecurityOrgGroupUpdate,
		Delete: resourceDcosSecurityOrgGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"gid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique group name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the newly created group",
			},
			"group_provider": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Provider for this group",
			},
		},
	}
}

func resourceDcosSecurityOrgGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	gid := d.Get("gid").(string)
	iamGroupCreate := dcos.IamGroupCreate{}
	if desc, ok := d.GetOk("description"); ok {
		iamGroupCreate.Description = desc.(string)
	}

	resp, err := client.IAM.CreateGroup(ctx, gid, iamGroupCreate)

	log.Printf("[TRACE] IAM.CreateGroup - %v", resp)

	if err != nil {
		return fmt.Errorf("Unable to create group: %s", err.Error())
	}

	d.SetId(gid)

	return nil
}

func resourceDcosSecurityOrgGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	gid := d.Id()

	group, resp, err := client.IAM.GetGroup(ctx, gid)

	log.Printf("[TRACE] IAM.GetGroup - %v", resp)

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] IAM.GetGroup - %s not found", gid)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Unable to read group: %s", err.Error())
	}

	d.Set("description", group.Description)
	d.SetId(group.Gid)

	return nil
}

func resourceDcosSecurityOrgGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	gid := d.Id()

	iamGroupUpdate := dcos.IamGroupUpdate{}
	if desc, ok := d.GetOk("description"); ok {
		iamGroupUpdate.Description = desc.(string)
	}

	_, err := client.IAM.UpdateGroup(ctx, gid, iamGroupUpdate)
	if err != nil {
		return err
	}

	return resourceDcosSecurityOrgGroupRead(d, meta)
}

func resourceDcosSecurityOrgGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	resp, err := client.IAM.DeleteGroup(ctx, d.Id())

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Unable to delete group: %s", err.Error())
	}

	d.SetId("")

	return nil
}
