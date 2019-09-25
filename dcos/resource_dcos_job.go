package dcos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
				Optional:    true,
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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Description:   "The command that is executed. This value is wrapped by Mesos via `/bin/sh -c ${job.cmd}`. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
				ConflictsWith: []string{"args"},
			},
			"args": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description:   "An array of strings that represents an alternative mode of specifying the command to run. This was motivated by safe usage of containerizer features like a custom Docker ENTRYPOINT. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
				ConflictsWith: []string{"cmd"}, // TODO: check if this really conflict withh cmd
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
				Type:          schema.TypeMap,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"ucr"},
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
			"ucr": {
				Type:          schema.TypeMap,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"docker"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The ucr repository image name.",
						},
					},
				},
			},
			"env": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "Environment variables",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The key/name of the variable",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    false,
							Description: "The value of the key/name",
						},
						"secret": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    false,
							Description: "The name of the secret.",
						},
					},
				},
			},
			"secrets": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    false,
				Description: "Any secrets that are necessary for the job",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"placement_constraint": {
				Type:     schema.TypeSet,
				Optional: true,
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
						"secret": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "Allow for volume secrets if using UCR.",
							ConflictsWith: []string{"docker"},
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
				Optional:     true,
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

	metronome_job, err := generateMetronomeJob(d, meta)
	if err != nil {
		return err
	}

	m_json, _ := json.Marshal(metronome_job)
	log.Printf("[TRACE] Pre-create MetronomeV1Job: %+v", metronome_job)
	log.Printf("[TRACE] MetronomeV1Job (json): %s", m_json)
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

	job, resp, err := getDCOSJobInfo(jobId, client, ctx)

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	setSchemaFromJob(d, &job)

	return nil
}

func setSchemaFromJob(d *schema.ResourceData, j *dcos.MetronomeV1Job) {
	d.Set("name", j.Id)
	d.Set("description", j.Description)

	d.Set("user", j.Run.User)
	// d.Set("labels", j.Run)
	d.Set("cmd", j.Run.Cmd)
	d.Set("args", j.Run.Args)

	// maybe this needs to be a loop
	d.Set("artifacts", j.Run.Artifacts)

	// TODO missing attributes
	if docker := j.Run.Docker; docker != nil {
		d.Set("docker.image", docker.Image)
	}

	// TODO missing attributes
	if ucr := j.Run.Ucr; ucr != nil {
		d.Set("ucr.image", ucr.Image)
	}

	if len(j.Run.Env) > 0 {
		envRes := make([]map[string]interface{}, 0, 1)

		for k, i := range j.Run.Env {
			entry := make(map[string]interface{})
			entry["key"] = k

			switch e := i.(type) {
			case map[string]string:
				entry["secret"] = e["secret"]
			case string:
				entry["value"] = e
			default:
				log.Printf("[WARNING] found key but no secret or value. Ignoring %v", i)
				continue
			}

			envRes = append(envRes, entry)
		}

		d.Set("env", envRes)
	}

	d.Set("secret", j.Run.Secrets)

	if constraints := j.Run.Placement.Constraints; constraints != nil {
		constraintsRes := make([]map[string]interface{}, 0)
		for _, constraint := range *constraints {
			c := make(map[string]interface{})

			c["attribute"] = constraint.Attribute
			c["operator"] = constraint.Operator
			c["value"] = constraint.Value

			constraintsRes = append(constraintsRes, c)
		}

		d.Set("placement_constraint", constraintsRes)
	}

	if restart := j.Run.Restart; restart != nil {
		d.Set("restart.active_deadline_seconds", restart.ActiveDeadlineSeconds)
		d.Set("restart.policy", restart.Policy)
	}

	if len(j.Run.Volumes) > 0 {
		vols := make([]map[string]interface{}, 0)
		for _, volume := range j.Run.Volumes {
			vols = append(vols, map[string]interface{}{
				"container_path": volume.ContainerPath,
				"host_path":      volume.HostPath,
				"mode":           volume.Mode,
				"secret":         volume.Secret,
			})
		}
		d.Set("volume", vols)
	}

	d.Set("cpus", j.Run.Cpus)

	d.Set("mem", j.Run.Mem)

	d.Set("disk", j.Run.Disk)

	d.Set("max_launch_delay", j.Run.MaxLaunchDelay)
}

func resourceDcosJobUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	jobId := d.Get("name").(string)

	// Perform read on "name" to confirm it actually exists...
	_, _, err := getDCOSJobInfo(jobId, client, ctx)
	if err != nil {
		return err
	}

	metronome_job, err := generateMetronomeJob(d, meta)
	if err != nil {
		return err
	}

	m_json, _ := json.Marshal(metronome_job)
	log.Printf("[TRACE] Pre-create MetronomeV1Job: %+v", metronome_job)
	log.Printf("[TRACE] MetronomeV1Job (json): %s", m_json)
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

func getDCOSJobInfo(jobId string, client *dcos.APIClient, ctx context.Context) (dcos.MetronomeV1Job, *http.Response, error) {
	log.Printf("[INFO] Attempting to read job info (%s)", jobId)

	mv1job, resp, err := client.Metronome.V1GetJob(ctx, jobId, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to create DCOS job %s", err)
		return dcos.MetronomeV1Job{}, resp, err
	}

	log.Printf("[INFO] DCOS job successfully retrieved (%s)", jobId)
	log.Printf("[TRACE] Metronome Job Response object: %+v", mv1job)

	return mv1job, resp, err
}

func generateMetronomeJob(d *schema.ResourceData, meta interface{}) (dcos.MetronomeV1Job, error) {
	var metronome_job dcos.MetronomeV1Job
	var metronome_job_run dcos.MetronomeV1JobRun
	var metronome_job_run_docker dcos.MetronomeV1JobRunDocker
	var metronome_job_run_ucr dcos.MetronomeV1JobRunUcr
	var metronome_job_artifacts []dcos.MetronomeV1JobRunArtifacts
	var metronome_job_volumes []dcos.MetronomeV1JobRunVolumes
	var metronome_job_restart dcos.MetronomeV1JobRunRestart
	var metronome_job_placement dcos.MetronomeV1JobRunPlacement
	var metronome_job_placement_constraint []dcos.MetronomeV1JobRunPlacementConstraints

	metronome_job.Id = d.Get("name").(string)
	metronome_job_run.Cpus = d.Get("cpus").(float64)
	metronome_job_run.Mem = int64(d.Get("mem").(int))
	metronome_job_run.MaxLaunchDelay = int32(d.Get("max_launch_delay").(int))

	if cmd, ok := d.GetOk("cmd"); ok {
		metronome_job_run.Cmd = cmd.(string)
	}

	if desc, ok := d.GetOk("description"); ok {
		metronome_job.Description = desc.(string)
	}

	if args, ok := d.GetOk("args"); ok {
		metronome_job_run.Args = args.([]string)
	}

	if user, ok := d.GetOk("user"); ok {
		metronome_job_run.User = user.(string)
	}

	if disk, ok := d.GetOk("disk"); ok {
		metronome_job_run.Disk = int64(disk.(int))
	}

	// labels
	if l, ok := d.GetOk("labels"); ok {
		labels := l.(map[string]interface{})

		tmp_lbl := make(map[string]string)
		for k, v := range labels {
			tmp_lbl[k] = v.(string)
		}
		metronome_job.Labels = tmp_lbl
	} else {
		log.Printf("[TRACE] labels not set, skipping")
	}

	// env
	if e, ok := d.GetOk("env"); ok {
		env_config := e.(*schema.Set).List()

		log.Printf("[TRACE] env (config): %+v", env_config)
		env_map := make(map[string]interface{})

		for env := range env_config {
			a := env_config[env].(map[string]interface{})

			key, ok := a["key"].(string)
			if !ok {
				log.Print("[ERROR] env.key is not a string!")
			}

			value, ok := a["value"].(string)
			if !ok {
				log.Print("[ERROR] env.value is not a string!")
			}

			secret, ok := a["secret"].(string)
			if !ok {
				log.Print("[ERROR] env.secret is not a string!")
			}

			if key != "" {
				env_map[key] = value
			} else {
				log.Printf("[TRACE] env.key is not set")
			}

			if secret != "" {
				env_map[secret] = map[string]string{
					"secret": secret,
				}
			} else {
				log.Printf("[TRACE] env.secret is not set")
			}
		}

		log.Printf("[TRACE] env_map %+s", env_map)

		env_json, _ := json.Marshal(env_map)
		log.Printf("[TRACE] env_json %s", env_json)
		metronome_job_run.Env = env_map
	} else {
		log.Printf("[TRACE] env not set, skipping")
	}

	// secrets
	if s, ok := d.GetOk("secrets"); ok {
		secret_map := make(map[string]interface{})
		config_secret := s.(map[string]interface{})

		log.Printf("[TRACE] config_secret (config): %+v", config_secret)

		for k, v := range config_secret {
			secret_map[k] = map[string]string{
				"source": v.(string),
			}
		}

		log.Printf("[TRACE] env_secret: %+v", secret_map)

		secret_map_json, _ := json.Marshal(secret_map)
		log.Printf("[TRACE] secret_map_json %s", secret_map_json)
		metronome_job_run.Secrets = secret_map
	} else {
		log.Printf("[TRACE] secrets not set, skipping")
	}

	// placement_constraints
	if p, ok := d.GetOk("placement_constraint"); ok {
		placement_constraints := p.(*schema.Set).List()

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
	} else {
		log.Printf("[TRACE] placement_constraint not set, skipping")
	}

	// artifacts
	if a, ok := d.GetOk("artifacts"); ok {
		artifacts := a.(*schema.Set).List()

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

		metronome_job_run.Artifacts = metronome_job_artifacts
	} else {
		log.Printf("[TRACE] artifacts not set, skipping")
	}

	// docker
	if do, ok := d.GetOk("docker"); ok {
		docker_config := do.(map[string]interface{})
		log.Printf("[TRACE] docker (config): %+v", docker_config)

		image, ok := docker_config["image"].(string)
		if !ok {
			log.Print("[ERROR] docker.image is not a string!")
		}

		metronome_job_run_docker.Image = image
		metronome_job_run.Docker = &metronome_job_run_docker
	} else {
		log.Printf("[TRACE] docker not set, skipping")
	}

	// ucr
	if uc, ok := d.GetOk("ucr"); ok {
		ucr_config := uc.(map[string]interface{})
		log.Printf("[TRACE] ucr (config): %+v", ucr_config)

		image, ok := ucr_config["image"].(string)
		if !ok {
			log.Print("[ERROR] ucr.image is not a string!")
		}

		var metronome_job_run_ucr_image dcos.MetronomeV1JobRunUcrImage
		metronome_job_run_ucr_image.Id = image
		metronome_job_run_ucr.Image = metronome_job_run_ucr_image
		metronome_job_run.Ucr = &metronome_job_run_ucr
	} else {
		log.Printf("[TRACE] ucr not set, skipping")
	}

	// volumes
	if vo, ok := d.GetOk("volume"); ok {
		vols := vo.(*schema.Set).List()

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

			secret, ok := a["secret"].(string)
			if !ok {
				log.Print("[ERROR] volume.secret is not a string!")
			}

			tmp_secret_set := false
			for k, v := range metronome_job_run.Secrets {
				log.Printf("[TRACE] volume.secrets: %s:%s", k, v)

				if secret == k {
					tmp_secret_set = true
				}
			}

			if !tmp_secret_set {
				return dcos.MetronomeV1Job{}, fmt.Errorf("[ERROR] Expecting '%s' to be part of secrets configuration", secret)
			}

			metronome_job_volumes = append(metronome_job_volumes, dcos.MetronomeV1JobRunVolumes{
				ContainerPath: container_path,
				HostPath:      host_path,
				Mode:          mode,
				Secret:        secret,
			})
		}

		log.Printf("[TRACE] volumes (struct): %+v", metronome_job_volumes)

		metronome_job_run.Volumes = metronome_job_volumes
	} else {
		log.Printf("[TRACE] volume not set, skipping")
	}

	// restart
	if re, ok := d.GetOk("restart"); ok {
		restart_config := re.(map[string]interface{})
		log.Printf("[TRACE] restart (config): %+v", restart_config)

		policy, ok := restart_config["policy"].(string)
		if !ok {
			log.Print("[ERROR] restart.policy is not a string!")
		} else {
			metronome_job_restart.Policy = policy
		}

		// This is a hack; terraform is treating this TypeInt as a string
		var active_deadline_seconds int
		if restart_config["active_deadline_seconds"] != nil {
			var err2 error
			active_deadline_seconds, err2 = strconv.Atoi(restart_config["active_deadline_seconds"].(string))
			if err2 != nil {
				log.Print("[ERROR] restart.active_deadline_seconds is not an int!")
			}

			log.Printf("[TRACE] policy: %s, active_deadline_seconds: %d", policy, active_deadline_seconds)

			metronome_job_restart.ActiveDeadlineSeconds = int32(active_deadline_seconds)
		} else {
			log.Printf("[TRACE] active_deadline_seconds is nil")
		}

		log.Printf("[TRACE] Metronome restart object: %+v", metronome_job_restart)

		metronome_job_run.Restart = &metronome_job_restart
	} else {
		log.Printf("[TRACE] restart not set, skipping")
	}

	metronome_job.Run = metronome_job_run

	return metronome_job, nil
}
