package dcos

// Based on Code from
// https://github.com/rolyv/terraform-provider-marathon/blob/master/marathon/resource_marathon_app.go
//
// Forked out of:
// https://github.com/rolyv/terraform-provider-Marathon
// --> https://github.com/PTC-Global/terraform-provider-marathon
// ----> https://github.com/nicgrayson/terraform-provider-marathon
//
// Released under MIT License
//
// With contributions from:
// https://github.com/rubbish
// https://github.com/cihangirbesiktas
// https://github.com/DustinChaloupka
// https://github.com/wleese
// https://github.com/shawnsilva
// https://github.com/adamdecaf
// https://github.com/rolyv
// https://github.com/tuier
// https://github.com/knuckolls
// https://github.com/mariomarin
// https://github.com/nicgrayson
// https://github.com/sherzberg
// https://github.com/daniellockard
// https://github.com/fsniper
// https://github.com/rjeczalik
//
// To improve the user expirience on DC/OS we're integrating the marathon provider
// into our DC/OS provider. We'll also provide support for DC/OS related marathon
// features which are not useful for plain marathon.
//
// Users who are intrested into plain marathon support should have a look at the
// above mentioned root projects.
//
// Credits to all contributors who made the ground work for this marathon resource
//

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	marathon "github.com/gambol99/go-marathon"
)

var legacyStringRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func resourceMarathonApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceMarathonAppCreate,
		Read:   resourceMarathonAppRead,
		Update: resourceMarathonAppUpdate,
		Delete: resourceMarathonAppDelete,

		Schema: map[string]*schema.Schema{
			"dcos_framework": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan_path": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "v1/plan",
						},
						"timeout": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     600,
							Description: "Timeout in seconds to wait for a framework to complete deployment",
						},
						"is_framework": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"accepted_resource_roles": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"app_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"args": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"backoff_seconds": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
				ForceNew: false,
				Default:  1,
			},
			"backoff_factor": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
				ForceNew: false,
				Default:  1.15,
			},
			"cmd": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"constraints": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"operation": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameter": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"container": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"docker": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"force_pull_image": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"image": {
										Type:     schema.TypeString,
										Required: true,
									},
									"parameters": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"privileged": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"volumes": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"host_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"mode": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"external": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"options": {
													Type:     schema.TypeMap,
													Optional: true,
												},
												"name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"provider": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"persistent": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "root",
												},
												"size": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"max_size": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"port_mappings": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"host_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"service_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Default:  "tcp",
										Optional: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"network_names": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "DOCKER",
						},
					},
				},
			},
			"cpus": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  0.1,
				ForceNew: false,
			},
			"gpus": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ForceNew: false,
			},
			"disk": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  0,
				ForceNew: false,
			},
			"dependencies": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"env": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"fetch": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cache": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"executable": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"extract": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"health_checks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"protocol": {
							Type:     schema.TypeString,
							Default:  "HTTP",
							Optional: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"grace_period_seconds": {
							Type:     schema.TypeInt,
							Default:  300,
							Optional: true,
						},
						"interval_seconds": {
							Type:     schema.TypeInt,
							Default:  60,
							Optional: true,
						},
						"port_index": {
							Type:     schema.TypeInt,
							Default:  0,
							Optional: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"timeout_seconds": {
							Type:     schema.TypeInt,
							Default:  20,
							Optional: true,
						},
						"ignore_http_1xx": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"max_consecutive_failures": {
							Type:     schema.TypeInt,
							Default:  3,
							Optional: true,
						},
						"delay_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"instances": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: false,
			},
			"labels": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"mem": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  128,
				ForceNew: false,
			},
			"max_launch_delay_seconds": &schema.Schema{
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  300,
				ForceNew: false,
			},
			"networks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"CONTAINER", "CONTAINER/BRIDGE", "HOST"}, false),
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
			"require_ports": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: false,
			},
			"port_definitions": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:     schema.TypeString,
							Default:  "tcp",
							Optional: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(legacyStringRegexp, "Invalid name"),
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"secrets": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
			},
			"upgrade_strategy": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum_health_capacity": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
						"maximum_over_capacity": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"unreachable_strategy": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inactive_after_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"expunge_after_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"kill_selection": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "YOUNGEST_FIRST",
				ForceNew: false,
			},
			"user": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"executor": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			// many other "computed" values haven't been added.
		},
	}
}

type deploymentEvent struct {
	id    string
	state string
}

type marathonConf struct {
	config                   marathon.Config
	Client                   marathon.Marathon
	DefaultDeploymentTimeout time.Duration
}

func genMarathonConf(meta interface{}) (marathonConf, error) {
	client := meta.(*dcos.APIClient)

	marathonConfig := marathon.NewDefaultConfig()
	dcosConf := client.CurrentDCOSConfig()

	// FIXME: support mom by providing marathon path
	marathonConfig.URL = dcosConf.URL() + "/service/marathon"

	marathonConfig.HTTPClient = client.HTTPClient()
	marathonConfig.HTTPSSEClient = client.HTTPClient()

	marathonConfig.EventsTransport = marathon.EventsTransportSSE

	// FIXME: make this configurable for each app.
	// DefaultDeploymentTimeout: time.Duration(d.Get("deployment_timeout").(int)) * time.Second
	conf := marathonConf{
		config:                   marathonConfig,
		DefaultDeploymentTimeout: time.Duration(600) * time.Second,
	}

	log.Printf("[TRACE] - MarathonConfig ")

	if err := conf.loadAndValidate(); err != nil {
		return conf, err
	}

	return conf, nil
}

func (c *marathonConf) loadAndValidate() error {
	client, err := marathon.NewClient(c.config)
	c.Client = client
	return err
}

func readDeploymentEvents(meta *marathon.Marathon, c chan deploymentEvent, ready chan bool) error {
	client := *meta

	EventIDs := marathon.EventIDDeploymentSuccess | marathon.EventIDDeploymentFailed

	events, err := client.AddEventsListener(EventIDs)
	if err != nil {
		log.Fatalf("Failed to register for events, %s", err)
	}
	defer client.RemoveEventsListener(events)
	defer close(c)
	ready <- true

	for {
		select {
		case event := <-events:
			switch mEvent := event.Event.(type) {
			case *marathon.EventDeploymentSuccess:
				c <- deploymentEvent{mEvent.ID, event.Name}
			case *marathon.EventDeploymentFailed:
				c <- deploymentEvent{mEvent.ID, event.Name}
			}
		}
	}
}

func waitOnSuccessfulDeployment(c chan deploymentEvent, id string, timeout time.Duration) error {
	select {
	case dEvent := <-c:
		if dEvent.id == id {
			switch dEvent.state {
			case "deployment_success":
				return nil
			case "deployment_failed":
				return errors.New("Received deployment_failed event from marathon")
			}
		}
	case <-time.After(timeout):
		return errors.New("Deployment timeout reached. Did not receive any deployment events")
	}
	return nil
}

func resourceMarathonAppCreate(d *schema.ResourceData, meta interface{}) error {
	config, err := genMarathonConf(meta)
	if err != nil {
		return err
	}

	client := config.Client

	c := make(chan deploymentEvent, 100)
	ready := make(chan bool)
	go readDeploymentEvents(&client, c, ready)
	select {
	case <-ready:
	case <-time.After(60 * time.Second):
		return errors.New("Timeout getting an EventListener")
	}

	application := mapResourceToApplication(d)

	application, err = client.CreateApplication(application)
	if err != nil {
		log.Println("[ERROR] creating application", err)
		return err
	}
	d.Partial(true)
	d.SetId(application.ID)
	setSchemaFieldsForApp(application, d)

	for _, deploymentID := range application.DeploymentIDs() {
		err = waitOnSuccessfulDeployment(c, deploymentID.DeploymentID, config.DefaultDeploymentTimeout)
		if err != nil {
			log.Println("[ERROR] waiting for application for deployment", deploymentID, err)
			return err
		}
	}

	d.Partial(false)

	return resourceMarathonAppRead(d, meta)
}

func resourceMarathonAppRead(d *schema.ResourceData, meta interface{}) error {
	config, err := genMarathonConf(meta)
	if err != nil {
		return err
	}
	client := config.Client

	app, err := client.Application(d.Id())

	if err != nil {
		// Handle a deleted app
		if apiErr, ok := err.(*marathon.APIError); ok && apiErr.ErrCode == marathon.ErrCodeNotFound {
			d.SetId("")
			return nil
		}
		return err
	}

	if app != nil && app.ID == "" {
		d.SetId("")
	}

	if app != nil {
		appErr := setSchemaFieldsForApp(app, d)
		if appErr != nil {
			return appErr
		}
	}

	return nil
}

func setSchemaFieldsForApp(app *marathon.Application, d *schema.ResourceData) error {

	err := d.Set("app_id", app.ID)
	if err != nil {
		return errors.New("Failed to set app_id: " + err.Error())
	}

	d.SetPartial("app_id")

	err = d.Set("accepted_resource_roles", &app.AcceptedResourceRoles)
	if err != nil {
		return errors.New("Failed to set accepted_resource_roles: " + err.Error())
	}

	d.SetPartial("accepted_resource_roles")

	err = d.Set("args", app.Args)
	if err != nil {
		return errors.New("Failed to set args: " + err.Error())
	}

	d.SetPartial("args")

	err = d.Set("backoff_seconds", app.BackoffSeconds)
	if err != nil {
		return errors.New("Failed to set backoff_seconds: " + err.Error())
	}

	d.SetPartial("backoff_seconds")

	err = d.Set("backoff_factor", app.BackoffFactor)
	if err != nil {
		return errors.New("Failed to set backoff_factor: " + err.Error())
	}

	d.SetPartial("backoff_factor")

	err = d.Set("cmd", app.Cmd)
	if err != nil {
		return errors.New("Failed to set cmd: " + err.Error())
	}

	d.SetPartial("cmd")

	if app.Constraints != nil && len(*app.Constraints) > 0 {
		cMaps := make([]map[string]string, len(*app.Constraints))
		for idx, constraint := range *app.Constraints {
			cMap := make(map[string]string)
			cMap["attribute"] = constraint[0]
			cMap["operation"] = constraint[1]
			if len(constraint) > 2 {
				cMap["parameter"] = constraint[2]
			}
			cMaps[idx] = cMap
		}

		err := d.Set("constraints", cMaps)

		if err != nil {
			return errors.New("Failed to set contraints: " + err.Error())
		}
	} else {
		d.Set("constraints", nil)
	}
	d.SetPartial("constraints")

	if app.Networks != nil && len(*app.Networks) > 0 {
		networks := make([]map[string]interface{}, len(*app.Networks))
		for idx, network := range *app.Networks {
			nMap := make(map[string]interface{})
			if network.Mode != "" {
				nMap["mode"] = strings.ToUpper(string(network.Mode))
			}
			if network.Name != "" {
				nMap["name"] = network.Name
			}
			networks[idx] = nMap
		}
		err := d.Set("networks", networks)

		if err != nil {
			return errors.New("Failed to set networks: " + err.Error())
		}
	} else {
		d.Set("networks", nil)
	}

	d.SetPartial("health_checks")

	if app.Container != nil {
		container := app.Container

		containerMap := make(map[string]interface{})
		containerMap["type"] = container.Type

		if container.Docker != nil {
			docker := container.Docker
			dockerMap := make(map[string]interface{})
			containerMap["docker"] = []interface{}{dockerMap}

			dockerMap["image"] = docker.Image
			log.Println("DOCKERIMAGE: " + docker.Image)
			dockerMap["force_pull_image"] = *docker.ForcePullImage

			// Marathon 1.5 does not allow both docker.network and app.networks at the same config
			if app.Networks == nil {
				dockerMap["network"] = docker.Network
			}

			parameters := make([]map[string]string, len(*docker.Parameters))
			for idx, p := range *docker.Parameters {
				parameter := make(map[string]string, 2)
				parameter["key"] = p.Key
				parameter["value"] = p.Value
				parameters[idx] = parameter
			}

			if len(*docker.Parameters) > 0 {
				dockerMap["parameters"] = parameters
			}

			dockerMap["privileged"] = *docker.Privileged
		}

		if len(*container.Volumes) > 0 {
			volumes := make([]map[string]interface{}, len(*container.Volumes))
			for idx, volume := range *container.Volumes {
				volumeMap := make(map[string]interface{})
				volumeMap["container_path"] = volume.ContainerPath
				volumeMap["host_path"] = volume.HostPath
				volumeMap["mode"] = volume.Mode

				if volume.External != nil {
					external := make(map[string]interface{})
					external["name"] = volume.External.Name
					external["provider"] = volume.External.Provider
					external["options"] = *volume.External.Options
					externals := make([]interface{}, 1)
					externals[0] = external
					volumeMap["external"] = externals
				}
				if volume.Persistent != nil {
					persistent := make(map[string]interface{})
					persistent["type"] = volume.Persistent.Type
					persistent["size"] = volume.Persistent.Size
					persistent["max_size"] = volume.Persistent.MaxSize
					persistents := make([]interface{}, 1)
					persistents[0] = persistent
					volumeMap["persistent"] = persistents
				}

				volumes[idx] = volumeMap
			}

			containerMap["volumes"] = volumes
		} else {
			containerMap["volumes"] = nil
		}

		if container.PortMappings != nil && len(*container.PortMappings) > 0 {
			portMappings := make([]map[string]interface{}, len(*container.PortMappings))
			for idx, portMapping := range *container.PortMappings {
				pmMap := make(map[string]interface{})
				pmMap["container_port"] = portMapping.ContainerPort
				pmMap["host_port"] = portMapping.HostPort
				_, ok := d.GetOk("container.0.port_mappings." + strconv.Itoa(idx) + ".service_port")
				if ok {
					pmMap["service_port"] = portMapping.ServicePort
				}

				pmMap["protocol"] = portMapping.Protocol
				labels := make(map[string]string, len(*portMapping.Labels))
				for k, v := range *portMapping.Labels {
					labels[k] = v
				}
				pmMap["labels"] = labels
				pmMap["name"] = portMapping.Name
				if _, ok := d.GetOk("container.0.port_mappings." + strconv.Itoa(idx) + ".network_names.#"); ok {
					pmMap["network_names"] = portMapping.NetworkNames
				}
				portMappings[idx] = pmMap
			}

			containerMap["port_mappings"] = portMappings
		}

		containerList := make([]interface{}, 1)
		containerList[0] = containerMap
		err := d.Set("container", containerList)

		if err != nil {
			return errors.New("Failed to set container: " + err.Error())
		}
	}
	d.SetPartial("container")

	err = d.Set("cpus", app.CPUs)
	if err != nil {
		return errors.New("Failed to set cpus: " + err.Error())
	}

	d.SetPartial("cpus")

	err = d.Set("gpus", app.GPUs)
	if err != nil {
		return errors.New("Failed to set gpus: " + err.Error())
	}

	d.SetPartial("gpus")

	err = d.Set("disk", app.Disk)
	if err != nil {
		return errors.New("Failed to set disk: " + err.Error())
	}

	d.SetPartial("disk")

	if app.Dependencies != nil {
		err = d.Set("dependencies", &app.Dependencies)
		if err != nil {
			return errors.New("Failed to set dependencies: " + err.Error())
		}
	}

	d.SetPartial("dependencies")

	err = d.Set("env", app.Env)
	if err != nil {
		return errors.New("Failed to set env: " + err.Error())
	}

	d.SetPartial("env")

	if app.Fetch != nil && len(*app.Fetch) > 0 {
		fetches := make([]map[string]interface{}, len(*app.Fetch))
		for i, fetch := range *app.Fetch {
			fetches[i] = map[string]interface{}{
				"uri":        fetch.URI,
				"cache":      fetch.Cache,
				"executable": fetch.Executable,
				"extract":    fetch.Extract,
			}
		}
		err := d.Set("fetch", fetches)

		if err != nil {
			return errors.New("Failed to set fetch: " + err.Error())
		}
	} else {
		d.Set("fetch", nil)
	}

	d.SetPartial("fetch")

	if app.HealthChecks != nil && len(*app.HealthChecks) > 0 {
		healthChecks := make([]map[string]interface{}, len(*app.HealthChecks))
		for idx, healthCheck := range *app.HealthChecks {
			hMap := make(map[string]interface{})
			if healthCheck.Command != nil {
				hMap["command"] = []interface{}{map[string]string{"value": healthCheck.Command.Value}}
			}
			hMap["grace_period_seconds"] = healthCheck.GracePeriodSeconds
			if healthCheck.IgnoreHTTP1xx != nil {
				hMap["ignore_http_1xx"] = *healthCheck.IgnoreHTTP1xx
			}
			hMap["interval_seconds"] = healthCheck.IntervalSeconds
			if healthCheck.MaxConsecutiveFailures != nil {
				hMap["max_consecutive_failures"] = *healthCheck.MaxConsecutiveFailures
			}
			if healthCheck.Path != nil {
				hMap["path"] = *healthCheck.Path
			}
			if healthCheck.PortIndex != nil {
				hMap["port_index"] = *healthCheck.PortIndex
			}
			if healthCheck.Port != nil {
				hMap["port"] = *healthCheck.Port
			}

			hMap["protocol"] = healthCheck.Protocol
			hMap["timeout_seconds"] = healthCheck.TimeoutSeconds
			healthChecks[idx] = hMap
		}

		err := d.Set("health_checks", healthChecks)

		if err != nil {
			return errors.New("Failed to set health_checks: " + err.Error())
		}
	} else {
		d.Set("health_checks", nil)
	}

	d.SetPartial("health_checks")

	err = d.Set("instances", app.Instances)
	if err != nil {
		return errors.New("Failed to set instances: " + err.Error())
	}

	d.SetPartial("instances")

	err = d.Set("labels", app.Labels)
	if err != nil {
		return errors.New("Failed to set labels: " + err.Error())
	}

	d.SetPartial("labels")

	err = d.Set("mem", app.Mem)
	if err != nil {
		return errors.New("Failed to set mem: " + err.Error())
	}

	d.SetPartial("mem")

	err = d.Set("max_launch_delay_seconds", app.MaxLaunchDelaySeconds)
	if err != nil {
		return errors.New("Failed to set max_launch_delay_seconds: " + err.Error())
	}

	d.SetPartial("max_launch_delay_seconds")

	err = d.Set("require_ports", app.RequirePorts)
	if err != nil {
		return errors.New("Failed to set require_ports: " + err.Error())
	}

	d.SetPartial("require_ports")

	if app.PortDefinitions != nil && len(*app.PortDefinitions) > 0 {
		portDefinitions := make([]map[string]interface{}, len(*app.PortDefinitions))
		for idx, portDefinition := range *app.PortDefinitions {
			hMap := make(map[string]interface{})
			if portDefinition.Port != nil {
				hMap["port"] = *portDefinition.Port
			}
			if portDefinition.Protocol != "" {
				hMap["protocol"] = portDefinition.Protocol
			}
			if portDefinition.Name != "" {
				hMap["name"] = portDefinition.Name
			}
			if portDefinition.Labels != nil {
				hMap["labels"] = *portDefinition.Labels
			}
			portDefinitions[idx] = hMap
		}
		err := d.Set("port_definitions", portDefinitions)

		if err != nil {
			return errors.New("Failed to set port_definitions: " + err.Error())
		}
	} else {
		d.Set("port_definitions", nil)
	}

	if app.UpgradeStrategy != nil {
		usMap := make(map[string]interface{})
		usMap["minimum_health_capacity"] = *app.UpgradeStrategy.MinimumHealthCapacity
		usMap["maximum_over_capacity"] = *app.UpgradeStrategy.MaximumOverCapacity
		err := d.Set("upgrade_strategy", &[]interface{}{usMap})

		if err != nil {
			return errors.New("Failed to set upgrade_strategy: " + err.Error())
		}
	} else {
		d.Set("upgrade_strategy", nil)
	}
	d.SetPartial("upgrade_strategy")

	if app.UnreachableStrategy != nil {
		unrMap := make(map[string]interface{})
		if app.UnreachableStrategy.InactiveAfterSeconds != nil {
			unrMap["inactive_after_seconds"] = *app.UnreachableStrategy.InactiveAfterSeconds
		} else {
			unrMap["inactive_after_seconds"] = nil
		}

		if app.UnreachableStrategy.ExpungeAfterSeconds != nil {
			unrMap["expunge_after_seconds"] = *app.UnreachableStrategy.ExpungeAfterSeconds
		} else {
			unrMap["expunge_after_seconds"] = nil
		}

		err := d.Set("unreachable_strategy", &[]interface{}{unrMap})
		if err != nil {
			return errors.New("Failed to set unreachable_strategy: " + err.Error())
		}
	} else {
		d.Set("unreachable_strategy", nil)
	}
	d.SetPartial("unreachable_strategy")

	err = d.Set("kill_selection", app.KillSelection)
	if err != nil {
		return errors.New("Failed to set kill_selection: " + err.Error())
	}
	d.SetPartial("kill_selection")

	err = d.Set("user", app.User)
	if err != nil {
		return errors.New("Failed to set user: " + err.Error())
	}
	d.SetPartial("user")

	err = d.Set("executor", *app.Executor)
	if err != nil {
		return errors.New("Failed to set executor: " + err.Error())
	}
	d.SetPartial("executor")

	err = d.Set("version", app.Version)
	if err != nil {
		return errors.New("Failed to set version: " + err.Error())
	}
	d.SetPartial("version")

	return nil
}

func resourceMarathonAppUpdate(d *schema.ResourceData, meta interface{}) error {
	config, err := genMarathonConf(meta)
	if err != nil {
		return err
	}
	client := config.Client

	c := make(chan deploymentEvent, 100)
	ready := make(chan bool)
	go readDeploymentEvents(&client, c, ready)
	select {
	case <-ready:
	case <-time.After(60 * time.Second):
		return errors.New("Timeout getting an EventListener")
	}

	application := mapResourceToApplication(d)

	deploymentID, err := client.UpdateApplication(application, false)
	if err != nil {
		return err
	}

	err = waitOnSuccessfulDeployment(c, deploymentID.DeploymentID, config.DefaultDeploymentTimeout)
	if err != nil {
		return err
	}

	return nil
}

func resourceMarathonAppDelete(d *schema.ResourceData, meta interface{}) error {
	config, err := genMarathonConf(meta)
	if err != nil {
		return err
	}
	client := config.Client

	_, err = client.DeleteApplication(d.Id(), false)
	if err != nil {
		return err
	}

	return nil
}

func mapResourceToApplication(d *schema.ResourceData) *marathon.Application {
	application := new(marathon.Application)

	if v, ok := d.GetOk("accepted_resource_roles.#"); ok {
		acceptedResourceRoles := make([]string, v.(int))

		for i := range acceptedResourceRoles {
			acceptedResourceRoles[i] = d.Get("accepted_resource_roles." + strconv.Itoa(i)).(string)
		}

		if len(acceptedResourceRoles) != 0 {
			application.AcceptedResourceRoles = acceptedResourceRoles
		}
	}

	if v, ok := d.GetOk("app_id"); ok {
		application.ID = v.(string)
	}

	if v, ok := d.GetOk("args.#"); ok {
		args := make([]string, v.(int))

		for i := range args {
			args[i] = d.Get("args." + strconv.Itoa(i)).(string)
		}

		if len(args) != 0 {
			application.Args = &args
		}
	}

	if v, ok := d.GetOk("backoff_seconds"); ok {
		value := v.(float64)
		application.BackoffSeconds = &value
	}

	if v, ok := d.GetOk("backoff_factor"); ok {
		value := v.(float64)
		application.BackoffFactor = &value
	}

	if v, ok := d.GetOk("cmd"); ok {
		value := v.(string)
		application.Cmd = &value
	}

	if v, ok := d.GetOk("constraints.#"); ok {
		constraints := make([][]string, v.(int))

		for i := range constraints {
			cMap := d.Get(fmt.Sprintf("constraints.%d", i)).(map[string]interface{})

			if cMap["parameter"] == "" {
				constraints[i] = make([]string, 2)
				constraints[i][0] = cMap["attribute"].(string)
				constraints[i][1] = cMap["operation"].(string)
			} else {
				constraints[i] = make([]string, 3)
				constraints[i][0] = cMap["attribute"].(string)
				constraints[i][1] = cMap["operation"].(string)
				constraints[i][2] = cMap["parameter"].(string)
			}
		}

		application.Constraints = &constraints
	} else {
		application.Constraints = nil
	}

	if v, ok := d.GetOk("container.0.type"); ok {
		container := new(marathon.Container)
		t := v.(string)

		container.Type = t

		if image, dockerOK := d.GetOk("container.0.docker.0.image"); dockerOK {
			docker := new(marathon.Docker)
			docker.Image = image.(string)

			if v, ok := d.GetOk("container.0.docker.0.force_pull_image"); ok {
				value := v.(bool)
				docker.ForcePullImage = &value
			}

			if v, ok := d.GetOk("container.0.docker.0.parameters.#"); ok {
				for i := 0; i < v.(int); i++ {
					paramMap := d.Get(fmt.Sprintf("container.0.docker.0.parameters.%d", i)).(map[string]interface{})
					docker.AddParameter(paramMap["key"].(string), paramMap["value"].(string))
				}
			}

			if v, ok := d.GetOk("container.0.docker.0.privileged"); ok {
				value := v.(bool)
				docker.Privileged = &value
			}

			container.Docker = docker
		}

		if v, ok := d.GetOk("container.0.volumes.#"); ok {
			volumes := make([]marathon.Volume, v.(int))

			for i := range volumes {
				volume := new(marathon.Volume)
				volumes[i] = *volume

				volumeMap := d.Get(fmt.Sprintf("container.0.volumes.%d", i)).(map[string]interface{})

				if val, ok := volumeMap["container_path"]; ok {
					volumes[i].ContainerPath = val.(string)
				}
				if val, ok := volumeMap["host_path"]; ok {
					volumes[i].HostPath = val.(string)
				}
				if val, ok := volumeMap["mode"]; ok {
					volumes[i].Mode = val.(string)
				}

				if volumeMap["external"] != nil {
					externalMap := d.Get(fmt.Sprintf("container.0.volumes.%d.external.0", i)).(map[string]interface{})
					if len(externalMap) > 0 {
						external := new(marathon.ExternalVolume)
						if val, ok := externalMap["name"]; ok {
							external.Name = val.(string)
						}
						if val, ok := externalMap["provider"]; ok {
							external.Provider = val.(string)
						}
						if val, ok := externalMap["options"]; ok {
							optionsMap := val.(map[string]interface{})
							options := make(map[string]string, len(optionsMap))

							for key, value := range optionsMap {
								options[key] = value.(string)
							}
							external.Options = &options
						}
						volumes[i].External = external
					}
				}

				if volumeMap["persistent"] != nil {
					persistentMap := d.Get(fmt.Sprintf("container.0.volumes.%d.persistent.0", i)).(map[string]interface{})
					if len(persistentMap) > 0 {
						persistent := new(marathon.PersistentVolume)
						if val, ok := persistentMap["type"]; ok {
							persistent.Type = marathon.PersistentVolumeType(val.(string))
						}
						if val, ok := persistentMap["size"]; ok {
							persistent.Size = val.(int)
						}
						if val, ok := persistentMap["max_size"]; ok {
							persistent.MaxSize = val.(int)
						}
						volumes[i].Persistent = persistent
					}
				}
			}
			container.Volumes = &volumes
		}

		if v, ok := d.GetOk("container.0.port_mappings.#"); ok {
			portMappings := make([]marathon.PortMapping, v.(int))

			for i := range portMappings {
				portMapping := new(marathon.PortMapping)
				portMappings[i] = *portMapping

				pmMap := d.Get(fmt.Sprintf("container.0.port_mappings.%d", i)).(map[string]interface{})

				if val, ok := pmMap["container_port"]; ok {
					portMappings[i].ContainerPort = val.(int)
				}
				if val, ok := pmMap["host_port"]; ok {
					portMappings[i].HostPort = val.(int)
				}
				if val, ok := pmMap["protocol"]; ok {
					portMappings[i].Protocol = val.(string)
				}
				if val, ok := pmMap["service_port"]; ok {
					portMappings[i].ServicePort = val.(int)
				}
				if val, ok := pmMap["name"]; ok {
					portMappings[i].Name = val.(string)
				}

				labelsMap := d.Get(fmt.Sprintf("container.0.port_mappings.%d.labels", i)).(map[string]interface{})
				labels := make(map[string]string, len(labelsMap))
				for key, value := range labelsMap {
					labels[key] = value.(string)
				}
				portMappings[i].Labels = &labels

				netNamesList := d.Get(fmt.Sprintf("container.0.port_mappings.%d.network_names", i)).([]interface{})
				netNames := make([]string, len(netNamesList))
				for index, value := range netNamesList {
					netNames[index] = value.(string)
				}

				portMappings[i].NetworkNames = &netNames
			}

			container.PortMappings = &portMappings
		}

		application.Container = container
	}

	if v, ok := d.GetOk("cpus"); ok {
		application.CPUs = v.(float64)
	}

	if v, ok := d.GetOk("gpus"); ok {
		value := v.(float64)
		application.GPUs = &value
	}

	if v, ok := d.GetOk("disk"); ok {
		value := v.(float64)
		application.Disk = &value
	}

	if v, ok := d.GetOk("dependencies.#"); ok {
		dependencies := make([]string, v.(int))

		for i := range dependencies {
			dependencies[i] = d.Get("dependencies." + strconv.Itoa(i)).(string)
		}

		if len(dependencies) != 0 {
			application.Dependencies = dependencies
		}
	}

	if v, ok := d.GetOk("env"); ok {
		envMap := v.(map[string]interface{})
		env := make(map[string]string, len(envMap))

		for k, v := range envMap {
			env[k] = v.(string)
		}

		application.Env = &env
	} else {
		env := make(map[string]string, 0)
		application.Env = &env
	}

	if v, ok := d.GetOk("fetch.#"); ok {
		fetch := make([]marathon.Fetch, v.(int))

		for i := range fetch {
			fetchMap := d.Get(fmt.Sprintf("fetch.%d", i)).(map[string]interface{})

			if val, ok := fetchMap["uri"].(string); ok {
				fetch[i].URI = val
			}
			if val, ok := fetchMap["cache"].(bool); ok {
				fetch[i].Cache = val
			}
			if val, ok := fetchMap["executable"].(bool); ok {
				fetch[i].Executable = val
			}
			if val, ok := fetchMap["extract"].(bool); ok {
				fetch[i].Extract = val
			}
		}

		application.Fetch = &fetch
	}

	if v, ok := d.GetOk("health_checks.#"); ok {
		healthChecks := make([]marathon.HealthCheck, v.(int))

		for i := range healthChecks {
			healthCheck := new(marathon.HealthCheck)
			mapStruct := d.Get("health_checks." + strconv.Itoa(i)).(map[string]interface{})

			commands := mapStruct["command"].([]interface{})
			if len(commands) > 0 {
				commandMap := commands[0].(map[string]interface{})
				healthCheck.Command = &marathon.Command{Value: commandMap["value"].(string)}
				healthCheck.Protocol = "COMMAND"
				if prop, ok := mapStruct["path"]; ok {
					prop := prop.(string)
					if prop != "" {
						healthCheck.Path = &prop
					}
				}
			} else {
				if prop, ok := mapStruct["path"]; ok {
					prop := prop.(string)
					if prop != "" {
						healthCheck.Path = &prop
					}
				}

				if prop, ok := mapStruct["port_index"]; ok {
					prop := prop.(int)
					healthCheck.PortIndex = &prop
				}

				if prop, ok := mapStruct["protocol"]; ok {
					healthCheck.Protocol = prop.(string)
				}
			}

			if prop, ok := mapStruct["port"]; ok {
				prop := prop.(int)
				if prop > 0 {
					healthCheck.Port = &prop
				}
			}

			if prop, ok := mapStruct["timeout_seconds"]; ok {
				healthCheck.TimeoutSeconds = prop.(int)
			}

			if prop, ok := mapStruct["grace_period_seconds"]; ok {
				healthCheck.GracePeriodSeconds = prop.(int)
			}

			if prop, ok := mapStruct["interval_seconds"]; ok {
				healthCheck.IntervalSeconds = prop.(int)
			}

			if prop, ok := mapStruct["ignore_http_1xx"]; ok {
				prop := prop.(bool)
				healthCheck.IgnoreHTTP1xx = &prop
			}

			if prop, ok := mapStruct["max_consecutive_failures"]; ok {
				prop := prop.(int)
				healthCheck.MaxConsecutiveFailures = &prop
			}

			healthChecks[i] = *healthCheck
		}

		application.HealthChecks = &healthChecks
	} else {
		application.HealthChecks = nil
	}

	if v, ok := d.GetOk("instances"); ok {
		v := v.(int)
		application.Instances = &v
	}

	if v, ok := d.GetOk("labels"); ok {
		labelsMap := v.(map[string]interface{})
		labels := make(map[string]string, len(labelsMap))

		for k, v := range labelsMap {
			labels[k] = v.(string)
		}

		application.Labels = &labels
	}

	if v, ok := d.GetOk("mem"); ok {
		v := v.(float64)
		application.Mem = &v
	}

	if v, ok := d.GetOk("max_launch_delay_seconds"); ok {
		v := v.(float64)
		application.MaxLaunchDelaySeconds = &v
	}

	if v, ok := d.GetOk("networks.0.mode"); ok {
		mode := v.(string)

		var networkMode marathon.PodNetworkMode
		name := ""
		switch strings.ToLower(mode) {
		case "host":
			networkMode = marathon.HostNetworkMode
		case "container/bridge":
			networkMode = marathon.BridgeNetworkMode
		case "container":
			if n, ok := d.GetOk("networks.0.name"); ok {
				name = n.(string)
			}

			networkMode = marathon.ContainerNetworkMode
		}

		application.SetNetwork(name, networkMode)
	}

	if v, ok := d.GetOk("require_ports"); ok {
		v := v.(bool)
		application.RequirePorts = &v
	}

	if v, ok := d.GetOk("port_definitions.#"); ok {
		portDefinitions := make([]marathon.PortDefinition, v.(int))

		for i := range portDefinitions {
			portDefinition := new(marathon.PortDefinition)
			mapStruct := d.Get("port_definitions." + strconv.Itoa(i)).(map[string]interface{})

			if prop, ok := mapStruct["port"]; ok {
				prop := prop.(int)
				portDefinition.Port = &prop
			}

			if prop, ok := mapStruct["protocol"]; ok {
				portDefinition.Protocol = prop.(string)
			}

			if prop, ok := mapStruct["name"]; ok {
				portDefinition.Name = prop.(string)
			}

			labelsMap := d.Get(fmt.Sprintf("port_definitions.%d.labels", i)).(map[string]interface{})
			labels := make(map[string]string, len(labelsMap))
			for key, value := range labelsMap {
				labels[key] = value.(string)
			}

			if len(labelsMap) > 0 {
				portDefinition.Labels = &labels
			}

			portDefinitions[i] = *portDefinition
		}

		application.PortDefinitions = &portDefinitions
	}

	upgradeStrategy := marathon.UpgradeStrategy{}

	if v, ok := d.GetOkExists("upgrade_strategy.0.minimum_health_capacity"); ok {
		f, ok := v.(float64)
		if ok {
			upgradeStrategy.MinimumHealthCapacity = &f
		}
	} else {
		f := 1.0
		upgradeStrategy.MinimumHealthCapacity = &f
	}

	if v, ok := d.GetOkExists("upgrade_strategy.0.maximum_over_capacity"); ok {
		f, ok := v.(float64)
		if ok {
			upgradeStrategy.MaximumOverCapacity = &f
		}
	} else {
		f := 1.0
		upgradeStrategy.MaximumOverCapacity = &f
	}

	if _, ok := d.GetOk("upgrade_strategy"); ok {
		application.SetUpgradeStrategy(upgradeStrategy)
	}

	unreachableStrategy := marathon.UnreachableStrategy{}
	if v, ok := d.GetOkExists("unreachable_strategy.0.inactive_after_seconds"); ok {
		f, ok := v.(float64)
		if ok {
			unreachableStrategy.InactiveAfterSeconds = &f
		}
	}

	if v, ok := d.GetOkExists("unreachable_strategy.0.expunge_after_seconds"); ok {
		f, ok := v.(float64)
		if ok {
			unreachableStrategy.ExpungeAfterSeconds = &f
		}
	}

	if v, ok := d.GetOk("unreachable_strategy"); ok {
		switch v.(type) {
		case string:
			application.UnreachableStrategy = nil
		default:
			application.SetUnreachableStrategy(unreachableStrategy)
		}
	}

	if v, ok := d.GetOk("kill_selection"); ok {
		v := v.(string)
		application.KillSelection = v
	}

	if v, ok := d.GetOk("user"); ok {
		v := v.(string)
		application.User = v
	}

	if v, ok := d.GetOk("secrets"); ok {
		secretsMap := v.(map[string]interface{})
		secrets := make(map[string]marathon.Secret, len(secretsMap))

		for k, v := range secretsMap {
			secrets[k] = marathon.Secret{k, v.(string)}
		}

		application.Secrets = &secrets
	}

	return application
}
