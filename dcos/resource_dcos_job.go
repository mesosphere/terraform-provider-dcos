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
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"artifacts": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
						"executable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"extract": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"cache": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"docker": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: false,
						},
					},
				},
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

	artifacts := d.Get("artifacts").(*schema.Set).List()

	log.Printf("[TRACE] artifacts (config): %+v", artifacts)

	for artifact := range artifacts {
		a := artifacts[artifact].(map[string]interface{})
		log.Printf("[TRACE] artifact (loop): %+v", a)

		uri, ok := a["uri"].(string)
		if !ok {
			log.Print("[ERROR] artifact.uri is not a string!")
		}

		extract, ok := a["extract"].(bool)
		if !ok {
			log.Print("[ERROR] artifact.extract is not a bool!")
		}

		executable, ok := a["executable"].(bool)
		if !ok {
			log.Print("[ERROR] artifact.executable is not a bool!")
		}

		cache, ok := a["cache"].(bool)
		if !ok {
			log.Print("[ERROR] artifact.cache is not a bool!")
		}

		metronome_job_artifacts = append(metronome_job_artifacts, dcos.MetronomeV1JobRunArtifacts{
			Uri:        uri,
			Extract:    extract,
			Executable: executable,
			Cache:      cache,
		})
	}

	log.Printf("[TRACE] artifacts (struct): %+v", metronome_job_artifacts)

	docker_config := d.Get("docker").(map[string]interface{})
	log.Printf("[TRACE] docker (config): %+v", docker_config)

	image, ok := docker_config["image"].(string)
	if !ok {
		log.Print("[ERROR] docker.image is not a string!")
	}

	metronome_job_run_docker.Image = image

	metronome_job_run.Artifacts = metronome_job_artifacts
	metronome_job_run.Docker = &metronome_job_run_docker
	metronome_job.Run = metronome_job_run

	log.Printf("[TRACE] Pre-create MetronomeV1Job: %+v", metronome_job)
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

	artifacts := d.Get("artifacts").(*schema.Set).List()

	log.Printf("[TRACE] artifacts (config): %+v", artifacts)

	for artifact := range artifacts {
		a := artifacts[artifact].(map[string]interface{})
		log.Printf("[TRACE] artifact (loop): %+v", a)

		uri, ok := a["uri"].(string)
		if !ok {
			log.Print("[ERROR] artifact.uri is not a string!")
		}

		extract, ok := a["extract"].(bool)
		if !ok {
			log.Print("[ERROR] artifact.extract is not a bool!")
		}

		executable, ok := a["executable"].(bool)
		if !ok {
			log.Print("[ERROR] artifact.executable is not a bool!")
		}

		cache, ok := a["cache"].(bool)
		if !ok {
			log.Print("[ERROR] artifact.cache is not a bool!")
		}

		metronome_job_artifacts = append(metronome_job_artifacts, dcos.MetronomeV1JobRunArtifacts{
			Uri:        uri,
			Extract:    extract,
			Executable: executable,
			Cache:      cache,
		})
	}

	log.Printf("[TRACE] artifacts (struct): %+v", metronome_job_artifacts)

	docker_config := d.Get("docker").(map[string]interface{})
	log.Printf("[TRACE] docker (config): %+v", docker_config)

	image, ok := docker_config["image"].(string)
	if !ok {
		log.Print("[ERROR] docker.image is not a string!")
	}

	metronome_job_run_docker.Image = image

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
		log.Printf("[ERROR] Failed to create DCOS job %s", err)
		return dcos.MetronomeV1Job{}, err
	}
	if resp.StatusCode != 200 {
		return dcos.MetronomeV1Job{}, fmt.Errorf("[ERROR] Expecting response code of 200 (job found), but received %d", resp.StatusCode)
	}

	log.Printf("[INFO] DCOS job successfully retrieved (%s)", jobId)
	log.Printf("[TRACE] Metronome Job Response object: %+v", mv1job)

	return mv1job, nil
}
