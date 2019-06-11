package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// Ensure that the ACL exists
	rid = strings.Replace(rid, "/", "%252F", -1)
	resp, err := client.IAM.CreateResourceACL(context.TODO(), rid, dcos.IamaclCreate{})
	if err != nil {
		if resp == nil || resp.StatusCode != 409 {
			return fmt.Errorf(
				"Unable to create resource ACL for '%s': %s",
				rid,
				err.Error(),
			)
		}
		log.Printf("permission '%s:%s' for user '%s' already exists", rid, action, uid)
	}

	// Grant permission
	resp, err = client.IAM.PermitResourceUserAction(ctx, rid, uid, action)
	log.Printf("[TRACE] PermitResourceUserAction - %v", resp.Request)
	if err != nil {
		return fmt.Errorf(
			"Unable to grant '%s' action on '%s' resource for user '%s': %s",
			action,
			rid,
			uid,
			err.Error(),
		)
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", uid, rid, action))
	return nil
}

func inPermissions(permissions dcos.IamUserPermissions, rid string, action string) bool {
	log.Printf("[TRACE] InPermission - %v", permissions)
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

func resourceDcosIAMGrantUserRead(d *schema.ResourceData, meta interface{}) error {
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
