package dcos

import (
	"fmt"
	"log"
	"time"

	"github.com/dcos/client-go/dcos"
	marathon "github.com/gambol99/go-marathon"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceDcosMarathonAPP() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosMarathonAPPCreate,
		Read:   resourceDcosMarathonAPPRead,
		Update: resourceDcosMarathonAPPUpdate,
		Delete: resourceDcosMarathonAPPDelete,
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
				Description: "Marathon AppID",
			},
			"cmd": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "Command to run",
			},
			// "container": {
			// 	Type:     schema.TypeList,
			// 	Optional: true,
			// 	Computed: true,
			// 	MaxItems: 1,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"type": {
			// 				Type:     schema.TypeString,
			// 				Default:  "MESOS",
			// 				Optional: true,
			// 			},
			// 		},
			// 	},
			// },
			"args": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Command arguments",
			},
			"instances": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Description: "How many instances to run",
			},
			"cpus": {
				Type:        schema.TypeFloat,
				Optional:    true,
				ForceNew:    false,
				Default:     0.1,
				Description: "Amount of CPUs to allocate",
			},
			"mem": {
				Type:        schema.TypeFloat,
				Optional:    true,
				ForceNew:    false,
				Default:     128.0,
				Description: "Amount of memory to allocate",
			},
			"disk": {
				Type:        schema.TypeFloat,
				Optional:    true,
				ForceNew:    false,
				Description: "Disk space to allocate",
			},
			"healthchecks": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "Healthchecks to use",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The protocol to use with this health check ()",
						},
					},
				},
			},

			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application status. e.g. Running",
			},
			"health": {
				Type:         schema.TypeString,
				Computed:     true,
				Description:  "Application health",
				ValidateFunc: validation.StringInSlice([]string{"MESOS_HTTP", "MESOS_HTTPS", "MESOS_TCP"}, false),
			},
		},
	}
}

func marathonApplicationStatus(application *marathon.Application) string {
	switch {
	case application == nil:
		return "Unknown"
	case len(application.Deployments) > 0:
		return "Deploying"
	case application.Instances != nil && *application.Instances == 0 && application.TasksRunning == 0:
		return "Suspended"
	case application.Instances != nil && *application.Instances > 0 && application.TasksRunning > 0:
		return "Running"
	default:
		return "Unknown"
	}
}

func marathonApplicationHealth(application *marathon.Application) string {
	switch {
	case application.Instances == nil:
		return "Destroyed"
	case application.TasksHealthy == application.TasksRunning && application.TasksRunning == *application.Instances:
		return "Healthy"
	case application.TasksRunning != *application.Instances:
		return "Scaling"
	default:
		return "Unknown"
	}
}

func AppStatusRefreshFunc(client marathon.Marathon, appID string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		app, err := client.Application(appID)
		if err != nil {
			return nil, "", err
		}

		status := marathonApplicationStatus(app)

		return app, status, nil
	}
}

func AppHealthRefreshFunc(client marathon.Marathon, appID string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		app, err := client.Application(appID)
		if err != nil {
			return nil, "", err
		}

		health := marathonApplicationHealth(app)

		return app, health, nil
	}
}

func getMarathonClient(meta interface{}) (marathon.Marathon, error) {
	client := meta.(*dcos.APIClient)

	config := marathon.NewDefaultConfig()
	marathonPath := client.CurrentConfig().BasePath + "/service/marathon"
	config.URL = marathonPath
	config.HTTPClient = client.HTTPClient()

	return marathon.NewClient(config)
}

func resourceDcosMarathonAPPCreate(d *schema.ResourceData, meta interface{}) error {
	marathonClient, err := getMarathonClient(meta)
	if err != nil {
		return fmt.Errorf("Error creating marathon client %v", err)
	}

	name := d.Get("name").(string)

	app := &marathon.Application{
		ID: name,
	}
	cmd := d.Get("cmd").(string)
	app.Command(cmd)

	if args, ok := d.GetOk("args"); ok {
		for _, arg := range args.(*schema.Set).List() {
			app.AddArgs(arg.(string))
		}
	}

	cpus := d.Get("cpus").(float64)
	app.CPU(cpus)

	mem := d.Get("mem").(float64)
	app.Memory(mem)

	if disk, ok := d.GetOk("disk"); ok {
		d := disk.(float64)
		app.Disk = &d
	}

	_, err = marathonClient.CreateApplication(app)

	d.SetId(name)

	d.Partial(true)

	instances := d.Get("instances").(int)
	if instances > 0 {
		_, err := marathonClient.ScaleApplicationInstances(name, instances, false)

		if err != nil {
			return err
		}

		waitForDeployment := true
		waitForDeploymentTimeout := 5 * time.Minute

		// TODO: maybe resource.StateChangeConf and .WaitForState is a better option
		// if deploying is would be part of the schema
		if waitForDeployment {

			stateConf := &resource.StateChangeConf{
				Pending:    []string{"Deploying"},
				Target:     []string{"Running"},
				Refresh:    AppStatusRefreshFunc(marathonClient, name, []string{}),
				Timeout:    waitForDeploymentTimeout,
				Delay:      10 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			_, err := stateConf.WaitForState()
			if err != nil {
				return err
			}
		}
	}

	return resourceDcosMarathonAPPRead(d, meta)
}

func resourceDcosMarathonAPPRead(d *schema.ResourceData, meta interface{}) error {
	marathonClient, err := getMarathonClient(meta)
	if err != nil {
		return err
	}

	appName := d.Id()
	app, err := marathonClient.Application(appName)
	log.Printf("[TRACE] Read - marathonClient.Application - appname \"%s\" app %v err %v", appName, app, err)

	if err != nil {
		if apiErr, ok := err.(*marathon.APIError); ok {
			switch apiErr.ErrCode {
			case marathon.ErrCodeNotFound:
				log.Printf("[DEBUG] Application %s not found", appName)
				d.SetId("")
				return nil
			}
		}

		return fmt.Errorf("Unknown Error - %v", err)
	}
	log.Printf("[TRACE] Found Application %v", app)

	d.Set("name", app.ID)
	d.SetId(app.ID)
	d.Set("cmd", app.Cmd)
	if a := app.Args; a != nil {
		args := make([]string, 0, len(*a))
		for _, arg := range *a {
			args = append(args, arg)
		}

		d.Set("args", args)
	} else {
		d.Set("args", nil)
	}

	d.Set("instances", app.Instances)

	d.Set("cpu", app.CPUs)
	d.Set("mem", *app.Mem)

	if disk := app.Disk; disk != nil {
		d.Set("disk", *disk)
	} else {
		d.Set("disk", nil)
	}

	status := marathonApplicationStatus(app)
	d.Set("status", status)

	health := marathonApplicationHealth(app)
	d.Set("health", health)

	return nil
}

func resourceDcosMarathonAPPUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()

	return resourceDcosMarathonAPPRead(d, meta)
}

func resourceDcosMarathonAPPDelete(d *schema.ResourceData, meta interface{}) error {
	marathonClient, err := getMarathonClient(meta)
	if err != nil {
		return err
	}

	appName := d.Get("name").(string)

	deployment, err := marathonClient.DeleteApplication(appName, false)

	if err != nil {
		if apiErr, ok := err.(*marathon.APIError); ok {
			switch apiErr.ErrCode {
			case marathon.ErrCodeNotFound:
				log.Printf("[DEBUG] Application %s not found", appName)
				d.SetId("")
				return nil
			}
		}

		return fmt.Errorf("Unknown Error - %v", err)
	}

	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		hasDeployment, err := marathonClient.HasDeployment(deployment.DeploymentID)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if hasDeployment {
			return resource.RetryableError(fmt.Errorf("Application still in deployment with DeploymentID %s", deployment.DeploymentID))
		}
		// finished
		return nil
	})

	return err
}
