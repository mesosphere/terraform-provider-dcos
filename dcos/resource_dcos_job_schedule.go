package dcos

import (
	"context"
	"fmt"
	"log"
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
			"dcos_job_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the job.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Unique identifier for the schedule.",
				ValidateFunc: validateRegexp("^[a-zA-Z0-9]+$"),
			},
			"cron": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				Description:  "Cron based schedule string",
				ValidateFunc: validateRegexp("^[0-9\\*\\/,]{1,5} [0-9\\*\\/,\\-]{1,15} [0-9\\*\\/,]{1,5} [0-9\\*\\/,]{1,5} [0-9\\*\\/,]{1,5}$"),
			},
			"concurrency_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				Default:      "",
				Description:  "Defines the behavior if a job is started, before the current job has finished. ALLOW will launch a new job, even if there is an existing run.",
				ValidateFunc: validation.StringInSlice([]string{"ALLOW", ""}, false),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "Defines if the schedule is enabled or not.",
			},
			"starting_deadline_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Default:      1,
				Description:  "The number of seconds until the job is still considered valid to start.",
				ValidateFunc: validation.IntAtLeast(1),
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
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	var metronome_job_schedule dcos.MetronomeV1JobSchedule

	scheduleId := d.Get("name").(string)
	jobId := d.Get("dcos_job_id").(string)

	metronome_job_schedule.Id = scheduleId
	metronome_job_schedule.Cron = d.Get("cron").(string)

	if concurrency_policy, ok := d.GetOk("concurrency_policy"); ok {
		metronome_job_schedule.ConcurrencyPolicy = concurrency_policy.(string)
	}

	if enabled, ok := d.GetOk("enabled"); ok {
		metronome_job_schedule.Enabled = enabled.(bool)
	}

	if starting_deadline_seconds, ok := d.GetOk("starting_deadline_seconds"); ok {
		metronome_job_schedule.StartingDeadlineSeconds = int32(starting_deadline_seconds.(int))
	}

	if timezone, ok := d.GetOk("timezone"); ok {
		metronome_job_schedule.Timezone = timezone.(string)
	}

	log.Printf("[TRACE] job schedule struct (create): %+v", metronome_job_schedule)

	sched_struct, resp, err := client.Metronome.V1CreateJobSchedules(ctx, jobId, metronome_job_schedule)
	if err != nil {
		return err
	}

	log.Printf("[TRACE] schedule struct: %+v", sched_struct)

	if resp.StatusCode != 201 {
		return fmt.Errorf("[ERROR] Expecting response code of 201 (schedule created), but received %d", resp.StatusCode)
	}

	d.SetId(scheduleId)

	return nil
}

func resourceDcosJobScheduleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	scheduleId := d.Get("name").(string)
	jobId := d.Get("dcos_job_id").(string)

	job_schedule, resp, err := client.Metronome.V1GetJobSchedulesByScheduleId(ctx, jobId, scheduleId)
	if err != nil {
		return err
	}

	log.Printf("[TRACE] job_schedule (read): %+v", job_schedule)

	if resp.StatusCode != 200 {
		return fmt.Errorf("[ERROR] Expecting response code of 200 (schedule retreived), but received %d", resp.StatusCode)
	}

	return nil
}

func resourceDcosJobScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	var metronome_job_schedule dcos.MetronomeV1JobSchedule

	scheduleId := d.Get("name").(string)
	jobId := d.Get("dcos_job_id").(string)

	metronome_job_schedule.Id = scheduleId
	metronome_job_schedule.Cron = d.Get("cron").(string)

	if concurrency_policy, ok := d.GetOk("concurrency_policy"); ok {
		metronome_job_schedule.ConcurrencyPolicy = concurrency_policy.(string)
	}

	if enabled, ok := d.GetOk("enabled"); ok {
		metronome_job_schedule.Enabled = enabled.(bool)
	}

	if starting_deadline_seconds, ok := d.GetOk("starting_deadline_seconds"); ok {
		metronome_job_schedule.StartingDeadlineSeconds = int32(starting_deadline_seconds.(int))
	}

	if timezone, ok := d.GetOk("timezone"); ok {
		metronome_job_schedule.Timezone = timezone.(string)
	}

	_, err := client.Metronome.V1PutJobSchedulesByScheduleId(ctx, jobId, scheduleId, metronome_job_schedule)
	if err != nil {
		return err
	}

	d.SetId(scheduleId)

	return resourceDcosJobScheduleRead(d, meta)
}

func resourceDcosJobScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	scheduleId := d.Get("name").(string)
	jobId := d.Get("dcos_job_id").(string)

	_, err := client.Metronome.V1DeleteJobSchedulesByScheduleId(ctx, jobId, scheduleId)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
