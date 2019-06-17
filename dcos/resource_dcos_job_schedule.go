package dcos

import (
	"context"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosJobSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosJobScheduleCreate,
		Read:   resourceDcosJobScheduleRead,
		Update: resourceDcosJobScheduleUpdate,
		Delete: resourceDcosJobScheduleDelete,
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the job.",
			},
		},
	}
}

func resourceDcosJobScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	return nil
}

func resourceDcosJobScheduleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	return nil
}

func resourceDcosJobScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	return resourceDcosJobRead(d, meta)
}

func resourceDcosJobScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	d.SetId("")

	return nil
}
