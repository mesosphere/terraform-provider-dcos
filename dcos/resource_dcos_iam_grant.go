package dcos

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosIAMGrant() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosIAMGrantCreate,
		Read:   resourceDcosIAMGrantRead,
		Delete: resourceDcosIAMGrantDelete,
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

func resourceDcosIAMGrantCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Get("uid").(string)
	rid := d.Get("resource").(string)
	action := d.Get("action").(string)

	_, resp, _ := client.IAM.GetResourceACLs(ctx, rid)
	if resp.StatusCode != http.StatusOK {
		_, err := client.IAM.CreateResourceACL(ctx, rid, dcos.IamaclCreate{})
		if err != nil {
			return err
		}
	}

	_, err := client.IAM.PermitResourceUserAction(ctx, rid, uid, action)

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", uid, rid, action))
	return nil
}

func inPermissions(permissions dcos.IamUserPermissions, rid string, action string) bool {
	for _, perm := range permissions.Direct {
		if perm.Rid == rid {
			for _, permAction := range perm.Actions {
				if permAction.Name == action {
					return true
				}
			}
		}
	}
	return false
}

func resourceDcosIAMGrantRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Get("uid").(string)
	rid := d.Get("resource").(string)
	action := d.Get("action").(string)

	permissions, resp, err := client.IAM.GetUserPermissions(ctx, uid)

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	if inPermissions(permissions, rid, action) {
		d.SetId(fmt.Sprintf("%s-%s-%s", uid, rid, action))
		return nil
	}

	d.SetId("")
	return nil
}

func resourceDcosIAMGrantDelete(d *schema.ResourceData, meta interface{}) error {
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
