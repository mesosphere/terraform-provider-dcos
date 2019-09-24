package dcos

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosSecurityOrgUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosSecurityOrgUserCreate,
		Read:   resourceDcosSecurityOrgUserRead,
		Update: resourceDcosSecurityOrgUserUpdate,
		Delete: resourceDcosSecurityOrgUserDelete,
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the newly created service account",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Path to public key to use",
			},
		},
	}
}

func resourceDcosSecurityOrgUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Get("uid").(string)
	iamUserCreate, err := iamUserCreateFromResourceData(d)

	if err != nil {
		return err
	}

	resp, err := client.IAM.CreateUser(ctx, uid, iamUserCreate)

	log.Printf("[TRACE] IAM.CreateUser - %v", resp)

	if err != nil {
		return err
	}

	d.SetId(uid)

	return nil
}

func resourceDcosSecurityOrgUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Id()

	user, resp, err := client.IAM.GetUser(ctx, uid)

	log.Printf("[TRACE] IAM.GetUser - %v", resp)

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] IAM.GetUser - %s not found", uid)
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.Set("description", user.Description)
	d.SetId(uid)

	return nil
}

func resourceDcosSecurityOrgUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	uid := d.Id()
	iamUserUpdate := dcos.IamUserUpdate{}

	if password, passwordOK := d.GetOk("password"); d.HasChange("password") && passwordOK {
		iamUserUpdate.Password = password.(string)
		d.Set("password", password.(string))
	}

	if description, ok := d.GetOk("description"); ok {
		iamUserUpdate.Description = description.(string)
	}

	resp, err := client.IAM.UpdateUser(ctx, uid, iamUserUpdate)

	log.Printf("[TRACE] IAM.UpdateUser - %v", resp)

	if err != nil {
		return err
	}

	return resourceDcosSecurityOrgUserRead(d, meta)
}

func resourceDcosSecurityOrgUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	resp, err := client.IAM.DeleteUser(ctx, d.Id())

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
