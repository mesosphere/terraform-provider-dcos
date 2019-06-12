package dcos

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosJobCreate,
		Read:   resourceDcosJobRead,
		Update: resourceDcosJobUpdate,
		Delete: resourceDcosJobDelete,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"cmd": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"args": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"artifacts_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"artificats_exectuable": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"artifacts_extract": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"artifacts_cache": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"docker_image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"cpus": {
				Type:     schema.TypeFloat,
				Required: true,
				ForceNew: false,
			},
			"mem": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"disk": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceDcosJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	var metronome_job dcos.MetronomeV1Job
	var metronome_job_run dcos.MetronomeV1JobRun
	var metronome_job_run_docker dcos.MetronomeV1JobRunDocker
	var metronome_job_artifacts []dcos.MetronomeV1JobRunArtifacts

	metronome_job.Id = d.Get("name").(string)
	metronome_job.Description = d.Get("description").(string)
	metronome_job_run.Cpus = d.Get("cpus").(float64)
	metronome_job_run.Mem = int64(d.Get("mem").(int))
	metronome_job_run.Disk = int64(d.Get("disk").(int))

	if labels, ok := d.GetOk("labels"); ok {
		metronome_job.Labels = labels.(map[string]string)
	}

	if cmd, ok := d.GetOk("cmd"); ok {
		metronome_job_run.Cmd = cmd.(string)
	}

	if args, ok := d.GetOk("args"); ok {
		metronome_job_run.Args = args.([]string)
	}

	if artifacts_uri, ok := d.GetOk("artifacts_uri"); ok {
		metronome_job_artifacts[0].Uri = artifacts_uri.(string)
	}

	if artificats_exectuable, ok := d.GetOk("artificats_exectuable"); ok {
		metronome_job_artifacts[0].Executable = artificats_exectuable.(bool)
	}

	if artifacts_extract, ok := d.GetOk("artifacts_extract"); ok {
		metronome_job_artifacts[0].Extract = artifacts_extract.(bool)
	}

	if artifacts_cache, ok := d.GetOk("artifacts_cache"); ok {
		metronome_job_artifacts[0].Cache = artifacts_cache.(bool)
	}

	if docker_image, ok := d.GetOk("docker_image"); ok {
		metronome_job_run_docker.Image = docker_image.(string)
	}

	metronome_job_run.Artifacts = metronome_job_artifacts
	metronome_job_run.Docker = &metronome_job_run_docker
	metronome_job.Run = metronome_job_run

	log.Printf("[INFO] Creating DCOS Job: %s", d.Get("name").(string))

	resp_metronome_job, resp, err := client.Metronome.V1CreateJob(ctx, metronome_job)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("[ERROR] Expecting response code of 201 (job created), but received %d", resp.StatusCode)
	}

	log.Printf("[INFO] DCOS job successfully created (%s)", d.Get("name").(string))
	log.Printf("[TRACE] Metronome Job Response object: %+v", resp_metronome_job)

	d.SetId(d.Get("name").(string))

	return nil
}

func resourceDcosJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	jobId := d.Get("name").(string)

	_, err := getDCOSJobInfo(jobId, client, ctx)
	if err != nil {
		return err
	}

	return nil
}

func resourceDcosJobUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	jobId := d.Get("name").(string)

	// Perform read on "name" to confirm it actually exists...
	_, err := getDCOSJobInfo(jobId, client, ctx)
	if err != nil {
		return err
	}

	// Update the job
	var metronome_job dcos.MetronomeV1Job
	var metronome_job_run dcos.MetronomeV1JobRun
	var metronome_job_run_docker dcos.MetronomeV1JobRunDocker
	var metronome_job_artifacts []dcos.MetronomeV1JobRunArtifacts

	metronome_job.Id = d.Get("name").(string)
	metronome_job.Description = d.Get("description").(string)
	metronome_job_run.Cpus = d.Get("cpus").(float64)
	metronome_job_run.Mem = int64(d.Get("mem").(int))
	metronome_job_run.Disk = int64(d.Get("disk").(int))

	if labels, ok := d.GetOk("labels"); ok {
		metronome_job.Labels = labels.(map[string]string)
	}

	if cmd, ok := d.GetOk("cmd"); ok {
		metronome_job_run.Cmd = cmd.(string)
	}

	if args, ok := d.GetOk("args"); ok {
		metronome_job_run.Args = args.([]string)
	}

	if artifacts_uri, ok := d.GetOk("artifacts_uri"); ok {
		metronome_job_artifacts[0].Uri = artifacts_uri.(string)
	}

	if artificats_exectuable, ok := d.GetOk("artificats_exectuable"); ok {
		metronome_job_artifacts[0].Executable = artificats_exectuable.(bool)
	}

	if artifacts_extract, ok := d.GetOk("artifacts_extract"); ok {
		metronome_job_artifacts[0].Extract = artifacts_extract.(bool)
	}

	if artifacts_cache, ok := d.GetOk("artifacts_cache"); ok {
		metronome_job_artifacts[0].Cache = artifacts_cache.(bool)
	}

	if docker_image, ok := d.GetOk("docker_image"); ok {
		metronome_job_run_docker.Image = docker_image.(string)
	}

	metronome_job_run.Artifacts = metronome_job_artifacts
	metronome_job_run.Docker = &metronome_job_run_docker
	metronome_job.Run = metronome_job_run

	log.Printf("[INFO] Updating DCOS Job: %s", d.Get("name").(string))

	resp_metronome_job, resp, err := client.Metronome.V1UpdateJob(ctx, jobId, metronome_job)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("[ERROR] Expecting response code of 201 (job updated), but received %d", resp.StatusCode)
	}

	log.Printf("[INFO] DCOS job successfully updated (%s)", d.Get("name").(string))
	log.Printf("[TRACE] Metronome Job Response object: %+v", resp_metronome_job)

	d.SetId(d.Get("name").(string))

	return resourceDcosJobRead(d, meta)
}

func resourceDcosJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	jobId := d.Get("name").(string)

	log.Printf("[INFO] Attempting to delete (%s)", jobId)
	resp, err := client.Metronome.V1DeleteJob(ctx, jobId)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("[ERROR] Expecting response code of 200 (job deleted), but received %d", resp.StatusCode)
	}
	log.Printf("[INFO] DCOS job successfully deleted (%s)", jobId)

	d.SetId("")

	return nil
}

func getDCOSJobInfo(jobId string, client *dcos.APIClient, ctx context.Context) (dcos.MetronomeV1Job, error) {
	log.Printf("[INFO] Attempting to read job info (%s)", jobId)

	mv1job, resp, err := client.Metronome.V1GetJob(ctx, jobId, nil)
	if err != nil {
		return dcos.MetronomeV1Job{}, err
	}
	if resp.StatusCode != 200 {
		return dcos.MetronomeV1Job{}, fmt.Errorf("[ERROR] Expecting response code of 200 (job found), but received %d", resp.StatusCode)
	}

	log.Printf("[INFO] DCOS job successfully retrieved (%s)", jobId)
	log.Printf("[TRACE] Metronome Job Response object: %+v", mv1job)

	return mv1job, nil
}
