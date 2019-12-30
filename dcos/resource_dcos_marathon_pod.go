package dcos

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"marathon_service_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "service/marathon",
				ForceNew:    true,
				Description: "By default we use the default DC/OS marathon serivce: service/marathon. But to support marathon on marathon the service url can be schanged.",
			},
			"container": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"artifact": {
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
									"exec": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"command_shell": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
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
										Default:  0,
									},
									"gpus": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
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
						"env": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"exec": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"command_shell": {
										Type:     schema.TypeString,
										Optional: true,
									},
									// "overrideEntrypoint": {
									// 	Type:     schema.TypeBool,
									// 	Optional: true,
									// },
								},
							},
						},
						"image": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
							Description: "DC/OS secrets",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kind": {
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice([]string{"DOCKER", "APPC"}, false),
										Optional:     true,
									},
									"id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"force_pull": {
										Type:     schema.TypeBool,
										Default:  false,
										Optional: true,
									},
									"pull_config": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"secret": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"secret": {
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
										Default:  false,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// remove leading slash
					return strings.TrimLeft(old, "/") == strings.TrimLeft(new, "/")
				},
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
							Default:      "CONTAINER",
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
										Type:     schema.TypeFloat,
										Optional: true,
										// Description: "Secret name",
									},
									"backoff_factor": {
										Type:     schema.TypeFloat,
										Optional: true,
										// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
									},
									"max_launch_delay": {
										Type:     schema.TypeFloat,
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
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"secret": {
							Type:     schema.TypeString,
							Optional: true,
							// Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
						"persistent": {
							Type:        schema.TypeList,
							Optional:    true,
							ForceNew:    false,
							MaxItems:    1,
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

	pod.Name(d.Get("name").(string))

	if v, ok := d.GetOk("container"); ok {
		// containers := make([]marathon.PodContainer, v.(*schema.Set).Len())

		for _, c := range v.(*schema.Set).List() {
			container := marathon.NewPodContainer()
			val := c.(map[string]interface{})

			if ar, ok := val["artifact"]; ok {
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

					log.Printf("[TRACE] Marathon.POD schemaToMarathonPod Artifact %+v", artifact)

					container.AddArtifact(&artifact)
				}
			}

			// container.SetHealthCheck(healthcheck *marathon.PodHealthCheck)
			if hc, ok := val["health_check"]; ok {
				hci := hc.([]interface{})

				if len(hci) > 0 {
					h := hci[0].(map[string]interface{})
					healthcheck := marathon.NewPodHealthCheck()

					if hv, ok := h["exec"]; ok {
						hc := hv.([]interface{})

						if len(hc) > 0 {
							execHealthCheck := marathon.NewCommandHealthCheck()
							h = hc[0].(map[string]interface{})

							if v, ok := h["command_shell"]; ok {
								execHealthCheck.Command = marathon.PodCommand{}
								execHealthCheck.Command.Shell = v.(string)

							}
							healthcheck.Exec = execHealthCheck
						}
					}

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
						hh := hv.([]interface{})

						if len(hh) > 0 {
							h := hh[0].(map[string]interface{})
							httpHealthCheck := marathon.NewHTTPHealthCheck()

							if v, ok := h["path"]; ok {
								httpHealthCheck.SetPath(v.(string))
							}

							if v, ok := h["scheme"]; ok {
								httpHealthCheck.SetScheme(v.(string))
							}

							if v, ok := h["endpoint"]; ok {
								httpHealthCheck.SetEndpoint(v.(string))
							}

							healthcheck.SetHTTPHealthCheck(httpHealthCheck)
						}
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

			if en, ok := val["env"]; ok {
				for k, v := range en.(map[string]interface{}) {
					container.AddEnv(k, v.(string))
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

			if im, ok := val["exec"]; ok {
				res := im.([]interface{})

				if len(res) > 0 {
					h := res[0].(map[string]interface{})
					exec := marathon.PodExec{}

					if hv, ok := h["command_shell"]; ok {
						exec.Command = marathon.PodCommand{}
						exec.Command.Shell = hv.(string)
					}

					container.Exec = &exec
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

					if hv, ok := h["pull_config"]; ok {
						res := hv.([]interface{})
						if len(res) > 0 {
							hs := res[0].(map[string]interface{})
							if sec, ok := hs["secret"]; ok {
								pc := marathon.NewPullConfig(sec.(string))
								image.SetPullConfig(pc)
							}
						}
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

			if ep, ok := val["secret"]; ok {
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

			// Empty strings means not adding the container
			if container.Name != "" {
				pod.AddContainer(container)
			}
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

			if v, ok := n["mode"]; ok {
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
		pod.Scaling = &marathon.PodScalingPolicy{}

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
				backoff.SetBackoff(v.(float64))
			}

			if v, ok := d.GetOk("scheduling.0.backoff.0.backoff_factor"); ok {
				backoff.SetBackoffFactor(v.(float64))
			}

			if v, ok := d.GetOk("scheduling.0.backoff.0.max_launch_delay"); ok {
				backoff.SetMaxLaunchDelay(v.(float64))
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

			if secret, ok := vol["secret"].(string); ok && secret != "" {
				volume := marathon.NewPodVolumeSecret(name, secret)
				log.Printf("[TRACE] Marathon.POD schemaToMarathonPod Add secret volume %+v", volume)
				pod.AddVolume(volume)

				// ignore everything else if filebased secret
			} else if path, ok := vol["host"].(string); ok {
				volume := marathon.NewPodVolume(name, path)

				if pers, ok := vol["persistent"]; ok {
					if perslist := pers.([]interface{}); len(perslist) == 1 {
						p := perslist[0].(map[string]interface{})
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
				}
				log.Printf("[TRACE] Marathon.POD schemaToMarathonPod Add volume %+v", volume)
				pod.AddVolume(volume)
			}
		}
	}

	return pod, nil
}

func resourceDcosMarathonPodCreate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()
	mconf, err := genMarathonConf(d, meta)
	if err != nil {
		return err
	}

	pod, err := schemaToMarathonPod(d)
	if err != nil {
		return err
	}

	log.Printf("[TRACE] Marathon.POD Creating POD %+v", pod)

	_, err = mconf.Client.CreatePod(pod)
	if err != nil {
		return err
	}

	return resourceDcosMarathonPodRead(d, meta)
}

func resourceDcosMarathonPodRead(d *schema.ResourceData, meta interface{}) error {
	mconf, err := genMarathonConf(d, meta)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)

	pod, err := mconf.Client.Pod(name)
	if err != nil {
		return err
	}

	if pod == nil {
		return fmt.Errorf("Couldn't receive Pod")
	}

	d.SetId(name)

	if len(pod.Containers) > 0 {
		containers := make([]map[string]interface{}, 0)
		for _, container := range pod.Containers {
			c := make(map[string]interface{})

			if len(container.Artifacts) > 0 {
				a := make([]map[string]interface{}, 0)
				for _, art := range container.Artifacts {
					artifact := make(map[string]interface{})

					artifact["uri"] = art.URI
					artifact["dest_path"] = art.DestPath
					artifact["cache"] = art.Cache
					artifact["executable"] = art.Executable
					artifact["extract"] = art.Extract

					log.Printf("[TRACE] Marathon.POD Read Artifact %+v", artifact)

					a = append(a, artifact)
				}

				c["artifact"] = a
			}

			if h := container.HealthCheck; h != nil {
				healthchecks := make([]map[string]interface{}, 1)
				healthchecks[0] = make(map[string]interface{})
				healthchecks[0]["grace_period_seconds"] = *h.GracePeriodSeconds
				healthchecks[0]["interval_seconds"] = *h.IntervalSeconds
				healthchecks[0]["max_consecutive_failures"] = *h.MaxConsecutiveFailures
				healthchecks[0]["timeout_seconds"] = *h.TimeoutSeconds
				healthchecks[0]["delay_seconds"] = *h.DelaySeconds
				if h.HTTP != nil {
					http := make([]map[string]interface{}, 1)
					http[0] = make(map[string]interface{})
					http[0]["path"] = h.HTTP.Path
					http[0]["scheme"] = h.HTTP.Scheme
					http[0]["endpoint"] = h.HTTP.Endpoint
					healthchecks[0]["http"] = http
				}
				if h.Exec != nil {
					exec := make([]map[string]interface{}, 1)
					exec[0] = make(map[string]interface{})
					exec[0]["command_shell"] = h.Exec.Command
					healthchecks[0]["exec"] = exec
				}

				c["health_check"] = healthchecks
			}

			c["name"] = container.Name

			if container.Resources != nil {
				resources := make([]map[string]interface{}, 1)
				resources[0] = make(map[string]interface{})
				resources[0]["cpus"] = container.Resources.Cpus
				resources[0]["mem"] = container.Resources.Mem
				resources[0]["disk"] = container.Resources.Disk
				resources[0]["gpus"] = container.Resources.Gpus

				c["resources"] = resources
			}

			if len(container.Endpoints) > 0 {
				endpoints := make([]map[string]interface{}, 0)

				for _, e := range container.Endpoints {
					endpoint := make(map[string]interface{})

					endpoint["name"] = e.Name
					endpoint["container_port"] = e.ContainerPort
					endpoint["host_port"] = e.HostPort
					endpoint["protocol"] = e.Protocol
					endpoint["labels"] = e.Labels

					endpoints = append(endpoints, endpoint)
				}
				c["endpoints"] = endpoints
			}

			if container.Exec != nil {
				exec := make([]map[string]interface{}, 1)
				exec[0] = make(map[string]interface{})

				exec[0]["command_shell"] = container.Exec.Command.Shell

				c["exec"] = exec
			}

			if container.Image != nil {
				image := make([]map[string]interface{}, 1)
				image[0] = make(map[string]interface{})
				image[0]["kind"] = container.Image.Kind
				image[0]["id"] = container.Image.ID
				image[0]["force_pull"] = container.Image.ForcePull
				if container.Image.PullConfig != nil {
					log.Printf("[TRACE] Marathon.POD Read found PullConfig %+v", container.Image.PullConfig)
					pullConfig := make(map[string]interface{})
					pullConfig["secret"] = container.Image.PullConfig.Secret
					image[0]["pull_config"] = pullConfig
				}

				c["image"] = image
			}

			c["env"] = container.Env
			c["labels"] = container.Labels

			if len(container.Secrets) > 0 {
				secrets := make([]map[string]interface{}, 0)
				for _, v := range container.Secrets {
					secret := make(map[string]interface{})

					secret["env_var"] = v.EnvVar
					secret["source"] = v.Source

					secrets = append(secrets, secret)
				}
				c["secret"] = secrets
			}

			c["user"] = container.User

			if len(container.VolumeMounts) > 0 {
				volmounts := make([]map[string]interface{}, 0)

				for _, e := range container.VolumeMounts {
					volmount := make(map[string]interface{})

					volmount["name"] = e.Name
					volmount["mount_path"] = e.MountPath
					volmount["read_only"] = e.ReadOnly

					volmounts = append(volmounts, volmount)
				}

				c["volume_mounts"] = volmounts
			}

			if container.Lifecycle.KillGracePeriodSeconds != nil {
				lifecycle := make([]map[string]interface{}, 1)
				lifecycle[0] = make(map[string]interface{})
				lifecycle[0]["kill_grace_period_seconds"] = *container.Lifecycle.KillGracePeriodSeconds

				c["lifecycle"] = lifecycle
			}

			containers = append(containers, c)
		}

		d.Set("container", containers)
	}

	if pod.ExecutorResources != nil {
		d.Set("executor_resources.0.cpus", pod.ExecutorResources.Cpus)
		d.Set("executor_resources.0.mem", pod.ExecutorResources.Mem)
		d.Set("executor_resources.0.disk", pod.ExecutorResources.Disk)
	}

	d.Set("name", pod.ID)

	d.Set("labels", pod.Labels)

	if len(pod.Networks) > 0 {
		networks := make([]map[string]interface{}, 0)

		for _, n := range pod.Networks {
			network := make(map[string]interface{})

			if n.Name != "" {
				network["name"] = n.Name
			}
			switch n.Mode {
			case marathon.ContainerNetworkMode:
				network["mode"] = "CONTAINER"
			case marathon.BridgeNetworkMode:
				network["mode"] = "CONTAINER/BRIDGE"
			case marathon.HostNetworkMode:
				network["mode"] = "HOST"
			}

			if len(n.Labels) > 0 {
				network["labels"] = n.Labels
			}
			networks = append(networks, network)
		}

		d.Set("network", networks)
	}

	if pod.Scaling != nil {
		d.Set("scaling.0.kind", pod.Scaling.Kind)
		d.Set("scaling.0.instances", pod.Scaling.Instances)
		d.Set("scaling.0.max_instances", pod.Scaling.MaxInstances)
	}

	if pod.Scheduling != nil {
		if pod.Scheduling.Backoff != nil {

			if pod.Scheduling.Backoff.Backoff != nil {
				d.Set("scheduling.0.backoff.0.backoff", *pod.Scheduling.Backoff.Backoff)
			}

			if pod.Scheduling.Backoff.BackoffFactor != nil {
				d.Set("scheduling.0.backoff.0.backoff_factor", *pod.Scheduling.Backoff.BackoffFactor)
			}

			if pod.Scheduling.Backoff.MaxLaunchDelay != nil {
				d.Set("scheduling.0.backoff.0.max_launch_delay", *pod.Scheduling.Backoff.MaxLaunchDelay)
			}
		}

		if pod.Scheduling.Upgrade != nil {
			if pod.Scheduling.Upgrade.MinimumHealthCapacity != nil {
				d.Set("scheduling.0.upgrade.0.minimum_health_capacity", *pod.Scheduling.Upgrade.MinimumHealthCapacity)
			}
			if pod.Scheduling.Upgrade.MaximumOverCapacity != nil {
				d.Set("scheduling.0.upgrade.0.maximum_over_capacity", *pod.Scheduling.Upgrade.MaximumOverCapacity)
			}
		}

		d.Set("scheduling.0.kill_selection", pod.Scheduling.KillSelection)

		if pod.Scheduling.UnreachableStrategy != nil {
			if pod.Scheduling.UnreachableStrategy.InactiveAfterSeconds != nil {
				d.Set("scheduling.0.unreachable_strategy.0.inactive_after_seconds", *pod.Scheduling.UnreachableStrategy.InactiveAfterSeconds)
			}

			if pod.Scheduling.UnreachableStrategy.ExpungeAfterSeconds != nil {
				d.Set("scheduling.0.unreachable_strategy.0.expunge_after_seconds", *pod.Scheduling.UnreachableStrategy.ExpungeAfterSeconds)
			}
		}
	}

	if len(pod.Secrets) > 0 {
		secrets := make([]map[string]interface{}, 0)
		for k, v := range pod.Secrets {
			secret := make(map[string]interface{})
			secret["secret_name"] = k
			secret["env_var"] = v.EnvVar
			secret["source"] = v.Source

			secrets = append(secrets, secret)
		}

		d.Set("secrets", secrets)
	}

	d.Set("user", pod.User)

	if len(pod.Volumes) > 0 {
		volumes := make([]map[string]interface{}, 0)

		for _, n := range pod.Volumes {
			volume := make(map[string]interface{})

			volume["name"] = n.Name
			volume["host"] = n.Host
			volume["secret"] = n.Secret

			if n.Persistent != nil {
				pers := make([]map[string]interface{}, 1)
				pers[0] = make(map[string]interface{})

				switch n.Persistent.Type {
				case marathon.PersistentVolumeTypeRoot:
					pers[0]["type"] = "root"
				case marathon.PersistentVolumeTypePath:
					pers[0]["type"] = "path"
				case marathon.PersistentVolumeTypeMount:
					pers[0]["type"] = "mount"
				}

				pers[0]["size"] = n.Persistent.Size
				pers[0]["max_size"] = n.Persistent.MaxSize

				if n.Persistent.Constraints != nil {
					constraints := make([]map[string]interface{}, 0)

					for _, c := range *n.Persistent.Constraints {
						con := make(map[string]interface{})
						con["attribute"] = c[0]
						con["operation"] = c[1]
						if len(c) > 2 {
							con["parameter"] = c[2]
						}

						constraints = append(constraints, con)
					}

					pers[0]["constraints"] = constraints
				}

				volume["persistent"] = pers
			}
			volumes = append(volumes, volume)
		}

		d.Set("volume", volumes)
	}

	return nil
}

func resourceDcosMarathonPodUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()
	mconf, err := genMarathonConf(d, meta)
	if err != nil {
		return err
	}

	pod, err := schemaToMarathonPod(d)
	if err != nil {
		return err
	}

	_, err = mconf.Client.UpdatePod(pod, true)
	if err != nil {
		return err
	}

	return resourceDcosMarathonPodRead(d, meta)
}

func resourceDcosMarathonPodDelete(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()
	mconf, err := genMarathonConf(d, meta)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)

	dpl, err := mconf.Client.DeletePod(name, true)
	if err != nil {
		return err
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		ok, err := mconf.Client.HasDeployment(dpl.DeploymentID)
		if ok {
			return resource.RetryableError(fmt.Errorf("Delete still in progress"))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
