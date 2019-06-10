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
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"cmd": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"args": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"artifacts_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"artificats_exectuable": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"artifacts_extract": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"artifacts_cache": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"docker_image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpus": {
				Type:     schema.TypeFloat,
				Required: true,
				ForceNew: true,
			},
			"mem": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"disk": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDcosJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	metronome_job := dcos.MetronomeV1Job{
		Id:          d.Get("name").(string),
		Description: d.Get("description").(string),
		Run: dcos.MetronomeV1JobRun{
			Cmd:  d.Get("cmd").(string),
			Cpus: d.Get("cpus").(float64),
			Mem:  int64(d.Get("mem").(int)),
			Disk: int64(d.Get("disk").(int)),
			Docker: &dcos.MetronomeV1JobRunDocker{
				Image: d.Get("docker_image").(string),
			},
		},
	}

	resp_metronome_job, resp, err := client.Metronome.V1CreateJob(ctx, metronome_job)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp_metronome_job)
	fmt.Printf("%+v\n", resp)

	return nil
}

func resourceDcosJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	store := d.Get("store").(string)
	pathToSecret := d.Get("path").(string)

	secret, resp, err := client.Secrets.GetSecret(ctx, store, encodePath(pathToSecret), nil)

	log.Printf("[TRACE] Read - %v", resp)

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] Read - %s not found", pathToSecret)
		d.SetId("")
		return nil
	}

	if err != nil {
		return nil
	}

	d.Set("value", secret.Value)
	d.SetId(generateID(store, pathToSecret))

	return nil
}

func resourceDcosJobUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	secretsV1Secret := dcos.SecretsV1Secret{}
	secretsV1Secret.Value = d.Get("value").(string)

	pathToSecret := d.Get("path").(string)

	store := d.Get("store").(string)

	_, err := client.Secrets.UpdateSecret(ctx, store, encodePath(pathToSecret), secretsV1Secret)

	if err != nil {
		return err
	}

	return resourceDcosSecretRead(d, meta)
}

func resourceDcosJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	pathToSecret := d.Get("path").(string)
	store := d.Get("store").(string)

	resp, err := client.Secrets.DeleteSecret(ctx, store, pathToSecret)

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

//
// import (
// 	"time"
//
// 	"github.com/hashicorp/terraform/helper/schema"
// 	"github.com/mesosphere/dcos-api-go/dcos"
// 	"github.com/mesosphere/dcos-api-go/dcos/job"
// )
//
// func resourceDcosJob() *schema.Resource {
// 	return &schema.Resource{
// 		Create: resourceDcosJobCreate,
// 		Read:   resourceDcosJobRead,
// 		Update: resourceDcosJobUpdate,
// 		Delete: resourceDcosJobDelete,
// 		Importer: &schema.ResourceImporter{
// 			State: schema.ImportStatePassthrough,
// 		},
//
// 		SchemaVersion: 1,
// 		Timeouts: &schema.ResourceTimeout{
// 			Create: schema.DefaultTimeout(10 * time.Minute),
// 			Update: schema.DefaultTimeout(10 * time.Minute),
// 			Delete: schema.DefaultTimeout(20 * time.Minute),
// 		},
//
// 		Schema: map[string]*schema.Schema{
// 			"jobid": {
// 				Type:     schema.TypeString,
// 				Required: true,
// 				ForceNew: true,
// 			},
// 			"run": {
// 				Type:     schema.TypeList,
// 				Required: true,
// 				MaxItems: 1,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"disk": {
// 							Type:     schema.TypeFloat,
// 							Required: true,
// 						},
// 						"cpus": {
// 							Type:     schema.TypeFloat,
// 							Required: true,
// 						},
// 						"mem": {
// 							Type:     schema.TypeFloat,
// 							Required: true,
// 						},
// 						"cmd": {
// 							Type:     schema.TypeString,
// 							Optional: true,
// 						},
// 						"env": {
// 							Type:     schema.TypeMap,
// 							Optional: true,
// 						},
// 						"maxlaunchdelay": {
// 							Type:     schema.TypeInt,
// 							Optional: true,
// 						},
// 						"artifacts": {
// 							Type:     schema.TypeList,
// 							Optional: true,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"uri": {
// 										Type:     schema.TypeString,
// 										Required: true,
// 									},
// 									"cache": {
// 										Type:     schema.TypeBool,
// 										Optional: true,
// 									},
// 									"executable": {
// 										Type:     schema.TypeBool,
// 										Optional: true,
// 									},
// 									"extract": {
// 										Type:     schema.TypeBool,
// 										Optional: true,
// 									},
// 								},
// 							},
// 						},
// 						"volumes": {
// 							Type:     schema.TypeList,
// 							Optional: true,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"containerpath": {
// 										Type:     schema.TypeString,
// 										Required: true,
// 									},
// 									"hostpath": {
// 										Type:     schema.TypeString,
// 										Required: true,
// 									},
// 									"mode": {
// 										Type:     schema.TypeString,
// 										Required: true,
// 									},
// 								},
// 							},
// 						},
// 						"docker": {
// 							Type:     schema.TypeList,
// 							Optional: true,
// 							MaxItems: 1,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"image": {
// 										Type:     schema.TypeString,
// 										Required: true,
// 									},
// 								},
// 							},
// 						},
// 						"restart": {
// 							Type:     schema.TypeList,
// 							Optional: true,
// 							MaxItems: 1,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"activedeadlineseconds": {
// 										Type:     schema.TypeInt,
// 										Required: true,
// 									},
// 									"policy": {
// 										Type:     schema.TypeString,
// 										Required: true,
// 									},
// 								},
// 							},
// 						},
// 						"placement": {
// 							Type:     schema.TypeList,
// 							Optional: true,
// 							MaxItems: 1,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"constraints": {
// 										Type:     schema.TypeList,
// 										Optional: true,
// 										Elem: &schema.Resource{
// 											Schema: map[string]*schema.Schema{
// 												"attribute": {
// 													Type:     schema.TypeString,
// 													Required: true,
// 												},
// 												"operator": {
// 													Type:     schema.TypeString,
// 													Required: true,
// 												},
// 												"value": {
// 													Type:     schema.TypeString,
// 													Optional: true,
// 												},
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			"description": {
// 				Type:     schema.TypeString,
// 				Optional: true,
// 			},
// 			"labels": {
// 				Type:     schema.TypeMap,
// 				Optional: true,
// 			},
// 		},
// 	}
// }
//
// func schemaToJob(d *schema.ResourceData) *job.Job {
// 	jobdef := job.Job{
// 		ID:          d.Get("jobid").(string),
// 		Description: d.Get("description").(string),
// 		Labels:      d.Get("labels").(map[string]string),
// 	}
//
// 	if r, ok := d.GetOk("run"); ok {
// 		runData := r.(*schema.ResourceData)
// 		run := &job.Run{}
//
// 		run.Cpus = runData.Get("cpus").(float64)
// 		run.Disk = runData.Get("disk").(float64)
// 		run.Mem = runData.Get("mem").(float64)
//
// 		run.Cmd = runData.Get("cmd").(string)
// 		run.MaxLaunchDelay = runData.Get("maxlaunchdelay").(int)
// 		run.User = runData.Get("user").(string)
// 		run.Env = runData.Get("env").(map[string]string)
//
// 		if a, ok := runData.GetOk("args"); ok {
// 			var args []string
//
// 			for _, arg := range a.(*schema.Set).List() {
// 				args = append(args, arg.(string))
// 			}
//
// 			run.Args = args
// 		}
//
// 		if a, ok := runData.GetOk("artifacts"); ok {
// 			arts := a.(*schema.Set).List()
// 			for _, artI := range arts {
// 				art := artI.(map[string]interface{})
// 				artifact := job.Artifact{
// 					URI: art["uri"].(string),
// 				}
//
// 				if v, ok := art["cache"].(bool); ok {
// 					artifact.Cache = v
// 				}
//
// 				if v, ok := art["executable"].(bool); ok {
// 					artifact.Executable = v
// 				}
//
// 				if v, ok := art["extract"].(bool); ok {
// 					artifact.Extract = v
// 				}
// 			}
// 		}
//
// 		if a, ok := runData.GetOk("docker"); ok {
// 			var d job.Docker
// 			dockerData := a.(*schema.ResourceData)
// 			d.Image = dockerData.Get("image").(string)
//
// 			run.Docker = &d
// 		}
//
// 		if p, ok := runData.GetOk("placement"); ok {
// 			var placement job.Placement
// 			placementData := p.(*schema.ResourceData)
//
// 			c := placementData.Get("constraints").(*schema.Set).List()
//
// 			for _, constr := range c {
// 				constrData := constr.(map[string]string)
// 				constraint := job.Constraint{
// 					Attribute: constrData["attribute"],
// 					Operator:  constrData["operator"],
// 				}
//
// 				if val, ok := constrData["value"]; ok {
// 					constraint.Value = val
// 				}
//
// 				placement.Constraints = append(placement.Constraints, &constraint)
// 			}
//
// 			run.Placement = &placement
// 		}
//
// 		if r, ok := runData.GetOk("restart"); ok {
// 			var restart job.Restart
// 			rest := r.(map[string]interface{})
//
// 			if val, ok := rest["policy"].(string); ok {
// 				restart.Policy = val
// 			}
//
// 			if val, ok := rest["activedeadlineseconds"].(int); ok {
// 				restart.ActiveDeadlineSeconds = val
// 			}
//
// 			run.Restart = &restart
// 		}
//
// 		if a, ok := runData.GetOk("volumes"); ok {
// 			var volumes []*job.Volume
// 			vols := a.(*schema.Set).List()
// 			for _, volI := range vols {
// 				volData := volI.(map[string]interface{})
//
// 				volume := job.Volume{
// 					ContainerPath: volData["containerpath"].(string),
// 					HostPath:      volData["hostpath"].(string),
// 					Mode:          volData["mode"].(string),
// 				}
//
// 				volumes = append(volumes, &volume)
// 			}
// 		}
//
// 		jobdef.Run = run
// 	}
//
// 	return &jobdef
// }
//
// func resourceDcosJobCreate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*dcos.Client)
// 	//
//
// 	jobdef := schemaToJob(d)
//
// 	client.Job.CreateJob(jobdef)
// 	//
// 	// app, err := client.Marathon.MarathonClient.CreateApplication(&application)
// 	//
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	//
// 	// d.SetId(app.ID)
//
// 	return resourceDcosJobRead(d, meta)
// }
//
// func resourceDcosJobRead(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*dcos.Client)
//
// 	app, err := client.Marathon.MarathonClient.Application(d.Id())
//
// 	d.Set("cmd", app.Cmd)
// 	d.Set("instances", app.Instances)
// 	d.Set("cpus", app.CPUs)
// 	d.Set("disk", app.Disk)
// 	d.Set("mem", app.Mem)
//
// 	return err
// }
//
// func resourceDcosJobUpdate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*dcos.Client)
//
// 	app, err := client.Marathon.MarathonClient.Application(d.Id())
// 	if err != nil {
// 		return err
// 	}
//
// 	d.Set("cmd", app.Cmd)
//
// 	if d.HasChange("cmd") {
// 		cmd := d.Get("cmd").(string)
// 		app.Cmd = &cmd
// 	}
// 	if d.HasChange("instances") {
// 		instances := d.Get("instances").(int)
// 		app.Instances = &instances
// 	}
// 	if d.HasChange("cpus") {
// 		app.CPUs = d.Get("cpus").(float64)
// 	}
// 	d.Set("disk", app.Disk)
// 	if d.HasChange("disk") {
// 		disk := d.Get("disk").(float64)
// 		app.Disk = &disk
// 	}
//
// 	d.Set("mem", app.Mem)
// 	if d.HasChange("mem") {
// 		mem := d.Get("mem").(float64)
// 		app.Mem = &mem
// 	}
//
// 	_, err = client.Marathon.MarathonClient.UpdateApplication(app, false)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	return resourceDcosJobRead(d, meta)
// }
//
// func resourceDcosJobDelete(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*dcos.Client)
//
// 	_, err := client.Marathon.MarathonClient.DeleteApplication(d.Id(), false)
//
// 	return err
// }
