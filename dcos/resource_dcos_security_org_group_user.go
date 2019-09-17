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

func resourceDcosIAMGroupUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosIAMGroupUserCreate,
		Read:   resourceDcosIAMGroupUserRead,
		// Update: resourceDcosIAMGroupUserUpdate,
		Delete: resourceDcosIAMGroupUserDelete,
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
				Description: "Group ID to assign a User to",
			},
			"uid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Userid to be assigned into group",
			},
		},
	}
}

func resourceDcosIAMGroupUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	gid := d.Get("gid").(string)
	uid := d.Get("uid").(string)

	resp, err := client.IAM.CreateGroupUser(ctx, gid, uid)
	log.Printf("[TRACE] IAM.CreateGroupUser - %v", resp)

	if err != nil {
		return fmt.Errorf("Unable to add user %s to group %s: %s", uid, gid, err.Error())
	}

	d.SetId(dcosIAMGroupUsergenID(d))

	return nil
}

func dcosIAMGroupUserinUserArray(uid string, users dcos.IamGroupUsers) bool {

	for _, user := range users.Array {
		if user.User.Uid == uid {
			return true
		}
	}

	return false
}

func dcosIAMGroupUsergenID(d *schema.ResourceData) string {
	gid := d.Get("gid").(string)
	uid := d.Get("uid").(string)

	return fmt.Sprintf("%s-%s", gid, uid)
}

func resourceDcosIAMGroupUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	gid := d.Get("gid").(string)
	uid := d.Get("uid").(string)

	users, resp, err := client.IAM.GetGroupUsers(ctx, gid, &dcos.GetGroupUsersOpts{})

	if !dcosIAMGroupUserinUserArray(uid, users) {
		log.Printf("[INFO] IAM.GetGroupUsers - %s not in group %s", uid, gid)
		d.SetId("")
		return nil
	}

	log.Printf("[TRACE] IAM.GetGroupUsers - %v", resp)

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("[INFO] IAM.GetGroupUsers - %s not found", gid)
	}

	if err != nil {
		return fmt.Errorf("Unable to find user %s in group %s: %s", uid, gid, err.Error())
	}

	d.SetId(dcosIAMGroupUsergenID(d))

	return nil
}

func resourceDcosIAMGroupUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	gid := d.Get("gid").(string)
	uid := d.Get("uid").(string)

	resp, err := client.IAM.DeleteGroupUser(ctx, gid, uid)

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Unable to delete user %s from group %s: %s", uid, gid, err.Error())
	}

	d.SetId("")

	return nil
}
