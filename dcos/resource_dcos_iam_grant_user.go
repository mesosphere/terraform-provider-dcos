package dcos

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosIAMGrantUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosIAMGrantUserCreate,
		Read:   resourceDcosIAMGrantUserRead,
		Delete: resourceDcosIAMGrantUserDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"uid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the account is used by default",
			},

			"resource": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Grants to be used",
			},

			"action": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Grants to be used",
			},
		},
	}
}

func resourceDcosIAMGrantUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Get("uid").(string)
	rid := d.Get("resource").(string)
	action := d.Get("action").(string)

	err := iamEnsureRid(ctx, client, rid)

	if err != nil {
		return fmt.Errorf("EnsureRID error - %v", err)
	}

	resp, err := client.IAM.PermitResourceUserAction(ctx, rid, uid, action)
	log.Printf("[TRACE] PermitResourceUserAction - %v", resp)

	if err != nil {
		return fmt.Errorf("PermitResourceUserAction - %v", err)

	}

	d.SetId(fmt.Sprintf("%s-%s-%s", uid, rid, action))
	return nil
}

func resourceDcosIAMGrantUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Get("uid").(string)
	rid := d.Get("resource").(string)
	action := d.Get("action").(string)

	allowed, _, err := client.IAM.GetResourceUserAction(ctx, rid, uid, action)

	if err != nil {
		return err
	}

	if !allowed.Allowed {
		d.SetId("")
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", rid, uid, action))
	return nil
}

func resourceDcosIAMGrantUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Get("uid").(string)
	rid := d.Get("resource").(string)
	action := d.Get("action").(string)

	_, err := client.IAM.ForbidResourceUserAction(ctx, rid, uid, action)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
