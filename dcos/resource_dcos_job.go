package dcos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the job.",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The user to use to run the tasks on the agent.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "A description of this job.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    false,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Attaching metadata to jobs can be useful to expose additional information to other services.",
			},
			"cmd": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The command that is executed. This value is wrapped by Mesos via `/bin/sh -c ${job.cmd}`. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
			},
			"args": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "An array of strings that represents an alternative mode of specifying the command to run. This was motivated by safe usage of containerizer features like a custom Docker ENTRYPOINT. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
			},
			"artifacts": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URI to be fetched by Mesos fetcher module.",
						},
						"executable": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Set fetched artifact as executable.",
						},
						"extract": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Extract fetched artifact if supported by Mesos fetcher module.",
						},
						"cache": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Cache fetched artifact if supported by Mesos fetcher module.",
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
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The docker repository image name.",
						},
					},
				},
			},
			"env": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    false,
				Description: "Environment variables (non secret)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"env_secret": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    false,
				Description: "Enviroment variables (secrets)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"placement_constraint": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The attribute name for this constraint.",
						},
						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     false,
							Description:  "The operator for this constraint.",
							ValidateFunc: validation.StringInSlice([]string{"EQ", "LIKE", "UNLIKE"}, false),
						},
						"value": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    false,
							Description: "The value for this constraint.",
						},
					},
				},
			},
			"restart": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    false,
				Description: "Defines the behavior if a task fails.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_deadline_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     false,
							Default:      120,
							Description:  "If the job fails, how long should we try to restart the job. If no value is set, this means forever.",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"policy": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     false,
							Default:      "NEVER",
							Description:  "The policy to use if a job fails. NEVER will never try to relaunch a job. ON_FAILURE will try to start a job in case of failure.",
							ValidateFunc: validation.StringInSlice([]string{"NEVER", "ON_FAILURE"}, false),
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_path": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The path of the volume in the container.",
							ValidateFunc: validateRegexp("^/[^/].*$"),
						},
						"host_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The path of the volume on the host.",
						},
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Possible values are RO for ReadOnly and RW for Read/Write.",
							ValidateFunc: validation.StringInSlice([]string{"RO", "RW"}, false),
						},
					},
				},
			},
			"cpus": {
				Type:        schema.TypeFloat,
				Required:    true,
				ForceNew:    false,
				Description: "The number of CPU shares this job needs per instance. This number does not have to be integer, but can be a fraction.",
			},
			"mem": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     false,
				Description:  "The amount of memory in MB that is needed for the job per instance.",
				ValidateFunc: validation.IntAtLeast(32),
			},
			"disk": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     false,
				Description:  "How much disk space is needed for this job. This number does not have to be an integer, but can be a fraction.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"max_launch_delay": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Default:      3600,
				Description:  "The number of seconds until the job needs to be running. If the deadline is reached without successfully running the job, the job is aborted.",
				ValidateFunc: validation.IntAtLeast(1),
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
	var metronome_job_volumes []dcos.MetronomeV1JobRunVolumes
	var metronome_job_restart dcos.MetronomeV1JobRunRestart
	var metronome_job_placement dcos.MetronomeV1JobRunPlacement
	var metronome_job_placement_constraint []dcos.MetronomeV1JobRunPlacementConstraints

	metronome_job.Id = d.Get("name").(string)
	metronome_job.Description = d.Get("description").(string)
	metronome_job_run.Cpus = d.Get("cpus").(float64)
	metronome_job_run.Mem = int64(d.Get("mem").(int))
	metronome_job_run.Disk = int64(d.Get("disk").(int))
	metronome_job_run.MaxLaunchDelay = int32(d.Get("max_launch_delay").(int))

	if labels, ok := d.GetOk("labels"); ok {
		metronome_job.Labels = labels.(map[string]string)
	}

	if cmd, ok := d.GetOk("cmd"); ok {
		metronome_job_run.Cmd = cmd.(string)
	}

	if args, ok := d.GetOk("args"); ok {
		metronome_job_run.Args = args.([]string)
	}

	if user, ok := d.GetOk("user"); ok {
		metronome_job_run.User = user.(string)
	}

	// env
	env_config := d.Get("env").(map[string]interface{})
	log.Printf("[TRACE] env (config): %+v", env_config)

	env_map := make(map[string]interface{})
	for k, v := range env_config {
		env_map[k] = v.(string)
	}

	env_config_secret := d.Get("env_secret").(map[string]interface{})
	log.Printf("[TRACE] env_secret (config): %+v", env_config_secret)
	for k, v := range env_config_secret {
		env_map[k] = dcos.MetronomeV1EnvSecretValue{
			Secret: v.(string),
		}
	}

	env_json, _ := json.Marshal(env_map)
	log.Printf("[TRACE] env_map (json): %s", env_json)
	log.Printf("[TRACE] env_map %+s", env_map)

	metronome_job_run.Env = env_map

	// placement_constraints
	placement_constraints := d.Get("placement_constraint").(*schema.Set).List()

	log.Printf("[TRACE] placement_constraints (config): %+v", placement_constraints)

	for constraint := range placement_constraints {
		a := placement_constraints[constraint].(map[string]interface{})
		log.Printf("[TRACE] constrant (loop): %+v", a)

		attribute, ok := a["attribute"].(string)
		if !ok {
			log.Print("[ERROR] placement_constraint.attribute is not a string!")
		}

		operator, ok := a["operator"].(string)
		if !ok {
			log.Print("[ERROR] placement_constraint.operator is not a string!")
		}

		value, ok := a["value"].(string)
		if !ok {
			log.Print("[ERROR] placement_constraint.value is not a string!")
		}

		metronome_job_placement_constraint = append(metronome_job_placement_constraint, dcos.MetronomeV1JobRunPlacementConstraints{
			Attribute: attribute,
			Operator:  operator,
			Value:     value,
		})
	}

	log.Printf("[TRACE] placement_constraint (struct): %+v", metronome_job_placement_constraint)

	metronome_job_placement.Constraints = &metronome_job_placement_constraint
	metronome_job_run.Placement = &metronome_job_placement

	// artifacts
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

	// docker
	docker_config := d.Get("docker").(map[string]interface{})
	log.Printf("[TRACE] docker (config): %+v", docker_config)

	image, ok := docker_config["image"].(string)
	if !ok {
		log.Print("[ERROR] docker.image is not a string!")
	}

	metronome_job_run_docker.Image = image

	// volumes
	vols := d.Get("volume").(*schema.Set).List()

	log.Printf("[TRACE] volumes (config): %+v", vols)

	for vol := range vols {
		a := vols[vol].(map[string]interface{})
		log.Printf("[TRACE] volume (loop): %+v", a)

		container_path, ok := a["container_path"].(string)
		if !ok {
			log.Print("[ERROR] volume.container_path is not a string!")
		}

		host_path, ok := a["host_path"].(string)
		if !ok {
			log.Print("[ERROR] volume.host_path is not a string!")
		}

		mode, ok := a["mode"].(string)
		if !ok {
			log.Print("[ERROR] volume.mode is not a string!")
		}

		metronome_job_volumes = append(metronome_job_volumes, dcos.MetronomeV1JobRunVolumes{
			ContainerPath: container_path,
			HostPath:      host_path,
			Mode:          mode,
		})
	}

	log.Printf("[TRACE] volumes (struct): %+v", metronome_job_volumes)

	metronome_job_run.Volumes = metronome_job_volumes

	// restart
	restart_config := d.Get("restart").(map[string]interface{})
	log.Printf("[TRACE] restart (config): %+v", restart_config)

	policy, ok := restart_config["policy"].(string)
	if !ok {
		log.Print("[ERROR] restart.policy is not a string!")
	}

	// This is a hack; terraform is treating this TypeInt as a string
	active_deadline_seconds, err := strconv.Atoi(restart_config["active_deadline_seconds"].(string))
	if !ok {
		log.Print("[ERROR] restart.active_deadline_seconds is not an int!")
	}

	log.Printf("[TRACE] policy: %s, active_deadline_seconds: %d", policy, active_deadline_seconds)

	metronome_job_restart.Policy = policy
	metronome_job_restart.ActiveDeadlineSeconds = int32(active_deadline_seconds)

	log.Printf("[TRACE] Metronome restart object: %+v", metronome_job_restart)

	metronome_job_run.Restart = &metronome_job_restart
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
	var metronome_job_volumes []dcos.MetronomeV1JobRunVolumes
	var metronome_job_restart dcos.MetronomeV1JobRunRestart
	var metronome_job_placement dcos.MetronomeV1JobRunPlacement
	var metronome_job_placement_constraint []dcos.MetronomeV1JobRunPlacementConstraints

	metronome_job.Id = d.Get("name").(string)
	metronome_job.Description = d.Get("description").(string)
	metronome_job_run.Cpus = d.Get("cpus").(float64)
	metronome_job_run.Mem = int64(d.Get("mem").(int))
	metronome_job_run.Disk = int64(d.Get("disk").(int))
	metronome_job_run.MaxLaunchDelay = int32(d.Get("max_launch_delay").(int))

	if labels, ok := d.GetOk("labels"); ok {
		metronome_job.Labels = labels.(map[string]string)
	}

	if cmd, ok := d.GetOk("cmd"); ok {
		metronome_job_run.Cmd = cmd.(string)
	}

	if args, ok := d.GetOk("args"); ok {
		metronome_job_run.Args = args.([]string)
	}

	if user, ok := d.GetOk("user"); ok {
		metronome_job_run.User = user.(string)
	}

	// env
	env_config := d.Get("env").(map[string]interface{})
	log.Printf("[TRACE] env (config): %+v", env_config)

	env_map := make(map[string]interface{})
	for k, v := range env_config {
		env_map[k] = v.(string)
	}

	env_config_secret := d.Get("env_secret").(map[string]interface{})
	log.Printf("[TRACE] env_secret (config): %+v", env_config_secret)
	for k, v := range env_config_secret {
		env_map[k] = dcos.MetronomeV1EnvSecretValue{
			Secret: v.(string),
		}
	}

	log.Printf("[TRACE] env_map %+s", env_map)

	metronome_job_run.Env = env_map

	// placement_constraints
	placement_constraints := d.Get("placement_constraint").(*schema.Set).List()

	log.Printf("[TRACE] placement_constraints (config): %+v", placement_constraints)

	for constraint := range placement_constraints {
		a := placement_constraints[constraint].(map[string]interface{})
		log.Printf("[TRACE] constrant (loop): %+v", a)

		attribute, ok := a["attribute"].(string)
		if !ok {
			log.Print("[ERROR] placement_constraint.attribute is not a string!")
		}

		operator, ok := a["operator"].(string)
		if !ok {
			log.Print("[ERROR] placement_constraint.operator is not a string!")
		}

		value, ok := a["value"].(string)
		if !ok {
			log.Print("[ERROR] placement_constraint.value is not a string!")
		}

		metronome_job_placement_constraint = append(metronome_job_placement_constraint, dcos.MetronomeV1JobRunPlacementConstraints{
			Attribute: attribute,
			Operator:  operator,
			Value:     value,
		})
	}

	log.Printf("[TRACE] placement_constraint (struct): %+v", metronome_job_placement_constraint)

	metronome_job_placement.Constraints = &metronome_job_placement_constraint
	metronome_job_run.Placement = &metronome_job_placement

	// artifacts
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

	// docker
	docker_config := d.Get("docker").(map[string]interface{})
	log.Printf("[TRACE] docker (config): %+v", docker_config)

	image, ok := docker_config["image"].(string)
	if !ok {
		log.Print("[ERROR] docker.image is not a string!")
	}

	metronome_job_run_docker.Image = image

	// volumes
	vols := d.Get("volume").(*schema.Set).List()

	log.Printf("[TRACE] volumes (config): %+v", vols)

	for vol := range vols {
		a := vols[vol].(map[string]interface{})
		log.Printf("[TRACE] volume (loop): %+v", a)

		container_path, ok := a["container_path"].(string)
		if !ok {
			log.Print("[ERROR] volume.container_path is not a string!")
		}

		host_path, ok := a["host_path"].(string)
		if !ok {
			log.Print("[ERROR] volume.host_path is not a string!")
		}

		mode, ok := a["mode"].(string)
		if !ok {
			log.Print("[ERROR] volume.mode is not a string!")
		}

		metronome_job_volumes = append(metronome_job_volumes, dcos.MetronomeV1JobRunVolumes{
			ContainerPath: container_path,
			HostPath:      host_path,
			Mode:          mode,
		})
	}

	log.Printf("[TRACE] volumes (struct): %+v", metronome_job_volumes)

	metronome_job_run.Volumes = metronome_job_volumes

	// restart
	restart_config := d.Get("restart").(map[string]interface{})
	log.Printf("[TRACE] restart (config): %+v", restart_config)

	policy, ok := restart_config["policy"].(string)
	if !ok {
		log.Print("[ERROR] restart.policy is not a string!")
	}

	// This is a hack; terraform is treating this TypeInt as a string
	active_deadline_seconds, err := strconv.Atoi(restart_config["active_deadline_seconds"].(string))
	if !ok {
		log.Print("[ERROR] restart.active_deadline_seconds is not an int!")
	}

	log.Printf("[TRACE] policy: %s, active_deadline_seconds: %d", policy, active_deadline_seconds)

	metronome_job_restart.Policy = policy
	metronome_job_restart.ActiveDeadlineSeconds = int32(active_deadline_seconds)

	log.Printf("[TRACE] Metronome restart object: %+v", metronome_job_restart)

	metronome_job_run.Restart = &metronome_job_restart
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
