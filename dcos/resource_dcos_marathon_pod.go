package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	marathon "github.com/gambol99/go-marathon"
)

func resourceDcosMarathonPod() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosMarathonPodCreate,
		Read:   resourceDcosMarathonPodRead,
		Update: resourceDcosMarathonPodUpdate,
		Delete: resourceDcosMarathonPodDelete,
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
			"containers": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"artifacts": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"dest_path": {
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
						"health_check": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grace_period_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"max_consecutive_failures": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  false,
									},
									"timeout_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  false,
									},
									"delay_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  true,
									},
									"http": {
										Type:        schema.TypeList,
										Optional:    true,
										ForceNew:    false,
										MaxItems:    1,
										Description: "DC/OS secrets",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"scheme": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"endpoint": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"resources": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpus": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"mem": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"disk": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"gpus": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"endpoints": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"container_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"host_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"protocol": &schema.Schema{
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"labels": {
										Type:     schema.TypeMap,
										Optional: true,
									},
								},
							},
						},
						"image": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kind": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"force_pull": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"env": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"secrets": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"env_var": {
										Type:     schema.TypeString,
										Required: true,
									},
									"source": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"volume_mounts": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"mount_path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"read_only": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"lifecycle": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kill_grace_period_seconds": {
										Type:     schema.TypeFloat,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"executor_resources": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				MaxItems:    1,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpus": {
							Type:     schema.TypeFloat,
							Required: true,
							// Description: "Secret name",
						},
						"mem": {
							Type:     schema.TypeFloat,
							Required: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"disk": {
							Type:     schema.TypeFloat,
							Default:  0,
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
					},
				},
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"network": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							// Description: "Secret name",
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"CONTAINER", "CONTAINER/BRIDGE", "HOST"}, false),
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
					},
				},
			},
			"scaling": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				MaxItems:    1,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:     schema.TypeString,
							Optional: true,
							// Description: "Secret name",
						},
						"instances": {
							Type:     schema.TypeInt,
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"max_instances": {
							Type:     schema.TypeInt,
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
					},
				},
			},
			"scheduling": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				MaxItems:    1,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backoff": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"backoff": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "Secret name",
									},
									"backoff_factor": {
										Type:     schema.TypeFloat,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
									"max_launch_delay": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
								},
							},
						},
						"upgrade": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"minimum_health_capacity": {
										Type:     schema.TypeFloat,
										Optional: true,
										// Description: "Secret name",
									},
									"maximum_over_capacity": {
										Type:     schema.TypeFloat,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
								},
							},
						},
						"kill_selection": {
							Type:     schema.TypeString,
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"unreachable_strategy": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"inactive_after_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "Secret name",
									},
									"expunge_after_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
								},
							},
						},
					},
				},
			},
			"secrets": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"env_var": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"volume": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							// Description: "Secret name",
						},
						"host": {
							Type:     schema.TypeString,
							Required: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"persistent": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"root", "path", "mount"}, false),
										// Description: "Secret name",
									},
									"size": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
									"max_size": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
									"constraints": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"attribute": {
													Type:     schema.TypeString,
													Required: true,
												},
												"operation": {
													Type:     schema.TypeString,
													Required: true,
												},
												"parameter": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func schemaToMarathonPod(d *schema.ResourceData) (*marathon.Pod, error) {
	pod := marathon.NewPod()

	if v, ok := d.GetOk("containers"); ok {
		// containers := make([]marathon.PodContainer, v.(*schema.Set).Len())

		for _, c := range v.(*schema.Set).List() {
			container := marathon.NewPodContainer()
			val := c.(map[string]interface{})

			if ar, ok := val["artifacts"]; ok {
				for _, a := range ar.(*schema.Set).List() {
					art := a.(map[string]interface{})
					artifact := marathon.PodArtifact{}

					if av, ok := art["uri"]; ok {
						artifact.URI = av.(string)
					}

					if av, ok := art["dest_path"]; ok {
						artifact.DestPath = av.(string)
					}

					if av, ok := art["cache"]; ok {
						artifact.Cache = av.(bool)
					}

					if av, ok := art["executable"]; ok {
						artifact.Executable = av.(bool)
					}

					if av, ok := art["extract"]; ok {
						artifact.Extract = av.(bool)
					}

					container.AddArtifact(&artifact)
				}
			}

			// container.SetHealthCheck(healthcheck *marathon.PodHealthCheck)
			if hc, ok := val["health_check"]; ok {
				hci := hc.([]interface{})

				if len(hci) > 0 {
					h := hci[0].(map[string]interface{})
					healthcheck := marathon.NewPodHealthCheck()

					if hv, ok := h["grace_period_seconds"]; ok {
						healthcheck.SetGracePeriod(hv.(int))
					}

					if hv, ok := h["interval_seconds"]; ok {
						healthcheck.SetInterval(hv.(int))
					}

					if hv, ok := h["max_consecutive_failures"]; ok {
						healthcheck.SetMaxConsecutiveFailures(hv.(int))
					}

					if hv, ok := h["timeout_seconds"]; ok {
						healthcheck.SetTimeout(hv.(int))
					}

					if hv, ok := h["delay_seconds"]; ok {
						healthcheck.SetDelay(hv.(int))
					}

					if hv, ok := h["http"]; ok {
						hh := hv.([]map[string]interface{})
						httpHealthCheck := marathon.NewHTTPHealthCheck()

						if v, ok := hh[0]["path"]; ok {
							httpHealthCheck.SetPath(v.(string))
						}

						if v, ok := hh[0]["scheme"]; ok {
							httpHealthCheck.SetScheme(v.(string))
						}

						if v, ok := hh[0]["endpoint"]; ok {
							httpHealthCheck.SetEndpoint(v.(string))
						}

						healthcheck.SetHTTPHealthCheck(httpHealthCheck)
					}
					container.SetHealthCheck(healthcheck)
				}
			}

			if n, ok := val["name"]; ok {
				container.SetName(n.(string))
			}

			if re, ok := val["resources"]; ok {
				res := re.([]interface{})

				if len(res) > 0 {
					h := res[0].(map[string]interface{})
					resources := marathon.NewResources()

					if hv, ok := h["cpus"]; ok {
						resources.Cpus = hv.(float64)
					}

					if hv, ok := h["mem"]; ok {
						resources.Mem = hv.(float64)
					}

					if hv, ok := h["disk"]; ok {
						resources.Disk = hv.(float64)
					}

					if hv, ok := h["gpus"]; ok {
						resources.Gpus = int32(hv.(int))
					}

					container.Resources = resources
				}
			}

			if ep, ok := val["endpoints"]; ok {
				for _, res := range ep.(*schema.Set).List() {
					h := res.(map[string]interface{})
					endpoint := marathon.NewPodEndpoint()

					if hv, ok := h["name"]; ok {
						endpoint.SetName(hv.(string))
					}

					if hv, ok := h["container_port"]; ok {
						endpoint.SetContainerPort(hv.(int))
					}

					if hv, ok := h["host_port"]; ok {
						endpoint.SetHostPort(hv.(int))
					}

					if hv, ok := h["protocol"]; ok {
						for _, i := range hv.([]interface{}) {
							endpoint.AddProtocol(i.(string))
						}
					}

					if hv, ok := h["labels"]; ok {
						for k, v := range hv.(map[string]interface{}) {
							endpoint.Label(k, v.(string))
						}
					}

					container.AddEndpoint(endpoint)
				}
			}

			if im, ok := val["image"]; ok {
				res := im.([]interface{})

				if len(res) > 0 {
					h := res[0].(map[string]interface{})
					image := marathon.NewPodContainerImage()

					if hv, ok := h["kind"]; ok {
						switch hv.(string) {
						case "APPC":
							image.SetKind(marathon.ImageTypeAppC)
						default:
							image.SetKind(marathon.ImageTypeDocker)
						}
					}

					if hv, ok := h["id"]; ok {
						image.SetID(hv.(string))
					}

					if hv, ok := h["force_pull"]; ok {
						image.ForcePull = (hv.(bool))
					}

					container.SetImage(image)
				}
			}

			if hv, ok := val["labels"]; ok {
				for k, v := range hv.(map[string]interface{}) {
					container.AddLabel(k, v.(string))
				}
			}

			if n, ok := val["name"]; ok {
				container.SetName(n.(string))
			}

			if ep, ok := val["secrets"]; ok {
				for _, res := range ep.(*schema.Set).List() {
					h := res.(map[string]interface{})

					container.AddSecret(h["env_var"].(string), h["source"].(string))
				}
			}

			if n, ok := val["user"]; ok {
				container.SetUser(n.(string))
			}

			if ep, ok := val["volume_mounts"]; ok {
				for _, res := range ep.(*schema.Set).List() {
					h := res.(map[string]interface{})
					volumemount := marathon.NewPodVolumeMount(h["name"].(string), h["mount_path"].(string))

					if hv, ok := h["read_only"]; ok {
						volumemount.ReadOnly = hv.(bool)
					}

					container.AddVolumeMount(volumemount)
				}
			}

			if re, ok := val["lifecycle"]; ok {
				res := re.([]interface{})

				if len(res) > 0 {
					h := res[0].(map[string]interface{})
					lifecycle := marathon.PodLifecycle{}

					if hv, ok := h["kill_grace_period_seconds"]; ok {
						seconds := hv.(float64)
						lifecycle.KillGracePeriodSeconds = &seconds
					}

					container.SetLifecycle(lifecycle)
				}
			}

			pod.AddContainer(container)
		}
	}

	if _, ok := d.GetOk("executor_resources"); ok {
		executorResources := marathon.ExecutorResources{}
		if v, ok := d.GetOk("executor_resources.0.cpu"); ok {
			executorResources.Cpus = v.(float64)
		}

		if v, ok := d.GetOk("executor_resources.0.mem"); ok {
			executorResources.Mem = v.(float64)
		}

		if v, ok := d.GetOk("executor_resources.0.disk"); ok {
			executorResources.Disk = v.(float64)
		}

		pod.SetExecutorResources(&executorResources)
	}

	if v, ok := d.GetOk("id"); ok {
		pod.Name(v.(string))
	}

	if lv, ok := d.GetOk("labels"); ok {
		for k, v := range lv.(map[string]interface{}) {
			pod.AddLabel(k, v.(string))
		}
	}

	if val, ok := d.GetOk("network"); ok {
		for _, c := range val.(*schema.Set).List() {
			network := marathon.PodNetwork{}
			n := c.(map[string]interface{})
			if v, ok := n["name"]; ok {
				network.SetName(v.(string))
			}

			if v, ok := n["name"]; ok {
				switch v.(string) {
				case "CONTAINER":
					network.SetMode(marathon.ContainerNetworkMode)
				case "CONTAINER/BRIDGE":
					network.SetMode(marathon.BridgeNetworkMode)
				case "HOST":
					network.SetMode(marathon.HostNetworkMode)
				}
			}

			if lv, ok := n["labels"]; ok {
				for k, v := range lv.(map[string]interface{}) {
					network.Label(k, v.(string))
				}
			}
			pod.AddNetwork(&network)
		}
	}

	if _, ok := d.GetOk("scaling"); ok {

		if v, ok := d.GetOk("scaling.0.kind"); ok {
			pod.Scaling.Kind = v.(string)
		}

		if v, ok := d.GetOk("scaling.0.instances"); ok {
			pod.Scaling.Instances = v.(int)
		}

		if v, ok := d.GetOk("scaling.0.max_instances"); ok {
			pod.Scaling.MaxInstances = v.(int)
		}

	}

	if _, ok := d.GetOk("scheduling"); ok {
		scheduling := marathon.PodSchedulingPolicy{}

		if _, ok := d.GetOk("scheduling.0.backoff"); ok {
			backoff := marathon.PodBackoff{}

			if v, ok := d.GetOk("scheduling.0.backoff.0.backoff"); ok {
				backoff.SetBackoff(v.(int))
			}

			if v, ok := d.GetOk("scheduling.0.backoff.0.backoff_factor"); ok {
				backoff.SetBackoffFactor(v.(float64))
			}

			if v, ok := d.GetOk("scheduling.0.backoff.0.max_launch_delay"); ok {
				backoff.SetMaxLaunchDelay(v.(int))
			}

			scheduling.SetBackoff(&backoff)
		}

		if _, ok := d.GetOk("scheduling.0.upgrade"); ok {
			upgrade := marathon.PodUpgrade{}

			if v, ok := d.GetOk("scheduling.0.upgrade.0.minimum_health_capacity"); ok {
				upgrade.SetMinimumHealthCapacity(v.(float64))
			}

			if v, ok := d.GetOk("scheduling.0.upgrade.0.maximum_over_capacity"); ok {
				upgrade.SetMaximumOverCapacity(v.(float64))
			}

			scheduling.SetUpgrade(&upgrade)
		}

		if v, ok := d.GetOk("scheduling.0.kill_selection"); ok {
			scheduling.SetKillSelection(v.(string))
		}

		if _, ok := d.GetOk("scheduling.0.unreachable_strategy	"); ok {
			upgrade := marathon.PodUpgrade{}

			if v, ok := d.GetOk("scheduling.0.upgrade.0.minimum_health_capacity"); ok {
				upgrade.SetMinimumHealthCapacity(v.(float64))
			}

			if v, ok := d.GetOk("scheduling.0.upgrade.0.maximum_over_capacity"); ok {
				upgrade.SetMaximumOverCapacity(v.(float64))
			}

			scheduling.SetUpgrade(&upgrade)
		}

		pod.SetPodSchedulingPolicy(&scheduling)

	}

	if val, ok := d.GetOk("secrets"); ok {
		for _, c := range val.(*schema.Set).List() {
			var envVar, secretName, sourceName string
			v := c.(map[string]interface{})

			secretName = v["secret_name"].(string)

			if e, ok := v["env_var"]; ok {
				envVar = e.(string)
			}

			if e, ok := v["source"]; ok {
				sourceName = e.(string)
			}
			pod.AddSecret(envVar, secretName, sourceName)
		}
	}

	if v, ok := d.GetOk("user"); ok {
		pod.SetUser(v.(string))
	}

	if val, ok := d.GetOk("volume"); ok {
		for _, i := range val.(*schema.Set).List() {
			vol := i.(map[string]interface{})

			name := vol["name"].(string)
			path := vol["host"].(string)

			volume := marathon.NewPodVolume(name, path)

			if pers, ok := vol["persistent"]; ok {
				p := pers.(map[string]interface{})
				persistent := marathon.PersistentVolume{}

				if v, ok := p["type"]; ok {
					switch v.(string) {
					case "root":
						persistent.SetType(marathon.PersistentVolumeTypeRoot)
					case "path":
						persistent.SetType(marathon.PersistentVolumeTypePath)
					case "mount":
						persistent.SetType(marathon.PersistentVolumeTypeMount)
					}
				}

				if v, ok := p["size"]; ok {
					persistent.SetSize(v.(int))
				}

				if v, ok := p["max_size"]; ok {
					persistent.SetMaxSize(v.(int))
				}

				if v, ok := p["max_size"]; ok {
					persistent.SetMaxSize(v.(int))
				}

				if c, ok := p["constraints"]; ok {
					for _, i := range c.(*schema.Set).List() {
						pers := i.(map[string]interface{})
						contraint := make([]string, 0)
						if p, ok := pers["attribute"]; ok {
							contraint = append(contraint, p.(string))
						}

						if p, ok := pers["operation"]; ok {
							contraint = append(contraint, p.(string))
						}

						if p, ok := pers["parameter"]; ok {
							contraint = append(contraint, p.(string))
						}

						persistent.AddConstraint(contraint...)
					}
				}
				volume.SetPersistentVolume(&persistent)
			}
			pod.AddVolume(volume)
		}
	}

	return pod, nil
}

func resourceDcosMarathonPodCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()
	mconf, err := genMarathonConf(meta)
	if err != nil {
		return err
	}

	pod, err := schemaToMarathonPod(d)
	if err != nil {
		return err
	}

	mconf.Client.CreatePod(pod)
	return nil
}

func resourceDcosMarathonPodRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	store := d.Get("store").(string)
	pathToSecret := d.Get("path").(string)

	secret, resp, err := client.Secrets.GetSecret(ctx, store, encodePath(pathToSecret), nil)

	log.Printf("[TRACE] Read - %v", resp)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
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

func resourceDcosMarathonPodUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	secretsV1Secret := dcos.SecretsV1Secret{}
	secretsV1Secret.Value = d.Get("value").(string)

	pathToSecret := d.Get("path").(string)

	store := d.Get("store").(string)

	_, err := client.Secrets.UpdateSecret(ctx, store, encodePath(pathToSecret), secretsV1Secret)

	if err != nil {
		return fmt.Errorf("Unable to update secret: %s", err.Error())
	}

	return resourceDcosMarathonPodRead(d, meta)
}

func resourceDcosMarathonPodDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	pathToSecret := d.Get("path").(string)
	store := d.Get("store").(string)

	resp, err := client.Secrets.DeleteSecret(ctx, store, pathToSecret)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Unable to delete secret: %s", err.Error())
	}

	d.SetId("")
	return nil
}
