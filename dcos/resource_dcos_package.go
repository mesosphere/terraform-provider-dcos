package dcos

import (
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosPackage() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosPackageCreate,
		Read:   resourceDcosPackageRead,
		Update: resourceDcosPackageUpdate,
		Delete: resourceDcosPackageDelete,
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
				Required:    true,
				ForceNew:    false,
				Description: "Description of the newly created service account",
			},
			"public_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"secret"},
				Description:   "Path to public key to use",
			},
			"secret": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"public_key"},
				Sensitive:     true,
				Description:   "Passphrase to use",
			},
		},
	}
}

func resourceDcosPackageCreate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()

	return nil
}

func resourceDcosPackageRead(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()

	// localVarOptionals := &dcos.ServiceDescribeOpts{}
	// localVarOptionals.CosmosServiceDescribeV1Request.
	//
	// client.Cosmos.ServiceDescribe(ctx, localVarOptionals)

	return nil
}

func resourceDcosPackageUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()

	return resourceDcosPackageRead(d, meta)
}

func resourceDcosPackageDelete(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()

	return nil
}
