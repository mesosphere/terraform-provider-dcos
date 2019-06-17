package dcos

import (
	"context"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
			"cron": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "Cron based schedule string",
			},
			"concurrency_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				Description:  "Defines the behavior if a job is started, before the current job has finished. ALLOW will launch a new job, even if there is an existing run.",
				ValidateFunc: validation.StringInSlice([]string{"ALLOW"}, false),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Description: "Defines if the schedule is enabled or not.",
			},
			"starting_deadline_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Description:  "The number of seconds until the job is still considered valid to start.",
				ValidateFunc: validation.IntAtLeast(32),
			},
			"timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				Description:  "IANA based time zone string. See http://www.iana.org/time-zones for a list of available time zones.",
				ValidateFunc: validateRegexp("^[a-zA-Z]+/?[a-zA-Z]+$"),
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
