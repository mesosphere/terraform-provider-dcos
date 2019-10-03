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

func resourceDcosEdgeLBV2Pool() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosEdgeLBV2PoolCreate,
		Read:   resourceDcosEdgeLBV2PoolRead,
		Update: resourceDcosEdgeLBV2PoolUpdate,
		Delete: resourceDcosEdgeLBV2PoolDelete,
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
			"pool_healthcheck_grace_period": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Pool tasks healthcheck grace period (in seconds)",
			},
			"pool_healthcheck_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Pool tasks healthcheck interval (in seconds)",
			},
			"pool_healthcheck_max_fail": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Pool tasks healthcheck maximum number of consecutive failures before declaring as unhealthy",
			},
			"pool_healthcheck_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Maximum amount of time that Mesos will wait for the healthcheck container to finish executing",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The pool name",
			},
			"namespace": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The DC/OS space (sometimes also referred to as a \"group\").",
			},
			"role": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "slave_public",
				ForceNew:    false,
				Description: "Mesos role for load balancers. Defaults to \"slave_public\" so that load balancers will be run on public agents. Use \"*\" to run load balancers on private agents. Read more about Mesos roles at http://mesos.apache.org/documentation/latest/roles/",
			},
			"principal": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Mesos principal for pool framework authentication. If omitted or left blank, the service account used to install Edge-LB will be used if present",
			},
			"secret_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Service account secret name for pool framework authentication. If omitted or left blank, the service account used to install Edge-LB will be used if present",
			},
			"cpus": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "CPU requirements",
			},
			"mem": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Memory requirements (in MB)",
			},
			"disk": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Description: "Disk size (in MB)",
			},
			"count": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Number of load balancer instances in the pool",
			},
			"constraints": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Marathon style constraints for load balancer instance placement",
			},
			"ports": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Elem:        schema.TypeInt,
				Description: "Override ports to allocate for each load balancer instance. Defaults to {{haproxy.frontends[].bindPort}} and   {{haproxy.stats.bindPort}}. Use this field to pre-allocate all needed ports with or   without the frontends present. For example: [80, 443, 9090]. If the length of the ports array is not zero, only the   ports specified will be allocated by the pool scheduler",
			},
			"secrets": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "DC/OS secrets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Secret name",
						},
						"file": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "File name. The file \"myfile\" will be found at \"$SECRETS/myfile\"",
						},
					},
				},
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Environment variables to pass to tasks. Prefix with \"ELB_FILE_\" and it will be written to a file. For example, the contents of \"ELB_FILE_MYENV\" will be written to \"$ENVFILE/ELB_FILE_MYENV\"",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auto_certificate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Autogenerate a self-signed SSL/TLS certificate. It is not generated by default. It will be written to \"$AUTOCERT\"",
			},
			"virtual_networks": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "Virtual networks to join",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"labels": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Labels to pass to the virtual network plugin.",
						},
					},
				},
			},

			"haproxy": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    false,
				Description: "Virtual networks to join",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"frontends": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "Virtual networks to join",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Defaults to frontend_{{bindAddress}}_{{bindPort}}",
									},
									"bind_address": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Only use characters that are allowed in the frontend name. Known invalid frontend name characters include \"*\", \"[\", and \"]\"",
									},
									"bind_port": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The port (e.g. 80 for HTTP or 443 for HTTPS) that this frontend will bind to",
									},
									"bind_modifier": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Additional text to put in the bind field",
									},
									"certificates": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     schema.TypeString,
									},
									"redirect_to_https": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"except": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"host": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Match on host",
															},
															"path_beg": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Match on path",
															},
														},
													},
												},
											},
										},
									},
									"misc_strs": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        schema.TypeString,
										Description: "Additional template lines inserted before use_backend",
									},
								},
							},
						},
						"backends": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "Virtual networks to join",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The name of the virtual network to join.",
									},
									"protocol": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Protocol",
									},
									"rewrite_http": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"host": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The name of the virtual network to join.",
												},
												"path": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"from_path": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "The name of the virtual network to join.",
															},
															"to_path": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "The name of the virtual network to join.",
															},
														},
													},
												},
												"request": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"forwardfor": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Set X-Forwarded-For",
															},
															"x_forwarded_port": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Set X-Forwarded-Port",
															},
															"x_forwarded_proto_https_if_tls": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Set X-Forwarded-Port HTTPS if TLS",
															},
															"set_host_header": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Set Host header",
															},
															"rewrite_path": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Rewrite Path",
															},
														},
													},
												},
												"response": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"rewrite_location": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "The name of the virtual network to join.",
															},
														},
													},
												},
												"sticky": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"enabled": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "The name of the virtual network to join.",
															},
															"custom_str": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "The name of the virtual network to join.",
															},
														},
													},
												},
											},
										},
									},
									"balance": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Load balancing strategy. e.g. roundrobin, leastconn, etc.",
									},
									"custom_check": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"httpchk": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "The name of the virtual network to join.",
												},
												"httpchk_misc_str": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The name of the virtual network to join.",
												},
												"ssl_hello_chk": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "The name of the virtual network to join.",
												},
												"misc_str": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "The name of the virtual network to join.",
												},
											},
										},
									},
									"misc_strs": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        schema.TypeString,
										Description: "Additional template lines inserted before servers",
									},
									"services": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"marathon": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"service_id": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Marathon pod or application ID",
															},
															"service_id_pattern": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
															"container_name": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Marathon pod container name, optional unless using Marathon pods",
															},
															"container_name_pattern": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
														},
													},
												},
												"mesos": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"framework_name": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Mesos framework name",
															},
															"framework_name_pattern": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
															"framework_id": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Mesos framework ID",
															},
															"framework_id_pattern": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
															"task_name": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Mesos task name",
															},
															"task_name_pattern": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
															"task_id": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Mesos task ID",
															},
															"task_id_pattern": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
														},
													},
												},
												"endpoint": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"type": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "Mesos framework name",
															},
															"misc_str": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Append arbitrary string to add to the end of the \"server\" directive",
															},
															"check": {
																Type:     schema.TypeSet,
																Optional: true,
																ForceNew: false,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"enabled": {
																			Type:     schema.TypeBool,
																			Optional: true,
																			// Description: "Mesos framework name",
																		},
																		"custom_str": {
																			Type:     schema.TypeString,
																			Optional: true,
																			// Description: "Append arbitrary string to add to the end of the \"server\" directive",
																		},
																	},
																},
															},
															"address": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Server address override, can be used to specify a cluster internal address such as a VIP",
															},
															"port": {
																Type:     schema.TypeInt,
																Optional: true,
																// Description: "Mesos task name",
															},
															"port_name": {
																Type:     schema.TypeString,
																Optional: true,
																// Description: "The name of the virtual network to join.",
															},
															"all_ports": {
																Type:     schema.TypeBool,
																Optional: true,
																// Description: "Mesos task ID",
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
				},
			},
		},
	}
}

func edgelbV2PoolFromSchema(d *schema.ResourceData) (dcos.EdgelbV2Pool, error) {
	edgelbV2Pool := dcos.EdgelbV2Pool{}

	poolName := d.Get("name").(string)

	edgelbV2Pool.Name = poolName

	if v, ok := d.GetOk("pool_healthcheck_grace_period"); ok {
		edgelbV2Pool.PoolHealthcheckGracePeriod = v.(int32)
	}
	if v, ok := d.GetOk("pool_healthcheck_interval"); ok {
		edgelbV2Pool.PoolHealthcheckInterval = v.(int32)
	}
	if v, ok := d.GetOk("pool_healthcheck_max_fail"); ok {
		edgelbV2Pool.PoolHealthcheckMaxFail = v.(int32)
	}
	if v, ok := d.GetOk("pool_healthcheck_timeout"); ok {
		edgelbV2Pool.PoolHealthcheckTimeout = v.(int32)
	}
	if v, ok := d.GetOk("namespace"); ok {
		edgelbV2Pool.Namespace = v.(string)
	}
	if v, ok := d.GetOk("role"); ok {
		edgelbV2Pool.Role = v.(string)
	}
	if v, ok := d.GetOk("principal"); ok {
		edgelbV2Pool.Principal = v.(string)
	}
	if v, ok := d.GetOk("secret_name"); ok {
		edgelbV2Pool.SecretName = v.(string)
	}
	if v, ok := d.GetOk("cpus"); ok {
		edgelbV2Pool.Cpus = v.(float32)
	}
	if v, ok := d.GetOk("count"); ok {
		edgelbV2Pool.Count = v.(int32)
	}
	if v, ok := d.GetOk("constraints"); ok {
		edgelbV2Pool.Constraints = v.(string)
	}
	if v, ok := d.GetOk("ports"); ok {
		edgelbV2Pool.Ports = v.([]int32)
	}

	if v, ok := d.GetOk("secrets"); ok {
		secrets := make([]dcos.EdgelbV2PoolSecrets, 0)
		for i := range v.(*schema.Set).List() {
			val := v.([]map[string]interface{})
			secret := dcos.EdgelbV2PoolSecrets{}

			if value, ok := val[i]["secret"]; ok {
				secret.Secret = value.(string)
			}

			if value, ok := val[i]["file"]; ok {
				secret.File = value.(string)
			}
			secrets = append(secrets, secret)
		}

		edgelbV2Pool.Secrets = secrets
	}

	if v, ok := d.GetOk("environment_variables"); ok {
		edgelbV2Pool.EnvironmentVariables = v.(map[string]string)
	}

	if v, ok := d.GetOk("auto_certificate"); ok {
		edgelbV2Pool.AutoCertificate = v.(bool)
	}

	if v, ok := d.GetOk("virtual_networks"); ok {
		networks := make([]dcos.EdgelbV2PoolVirtualNetworks, 0)

		for i := range v.(*schema.Set).List() {
			val := v.([]map[string]interface{})
			network := dcos.EdgelbV2PoolVirtualNetworks{}

			if value, ok := val[i]["name"]; ok {
				network.Name = value.(string)
			}

			if value, ok := val[i]["labels"]; ok {
				network.Labels = value.(map[string]string)
			}

			networks = append(networks, network)
		}
		edgelbV2Pool.VirtualNetworks = networks
	}

	if _, ok := d.GetOk("haproxy"); ok {
		if v, ok := d.GetOk("haproxy.frontends"); ok {
			frontends := make([]dcos.EdgelbV2Frontend, 0)
			for i := range v.(*schema.Set).List() {
				val := v.([]map[string]interface{})
				frontend := dcos.EdgelbV2Frontend{}

				if value, ok := val[i]["name"]; ok {
					frontend.Name = value.(string)
				}
				if value, ok := val[i]["bind_port"]; ok {
					frontend.BindPort = value.(int32)
				}
				if value, ok := val[i]["bind_modifier"]; ok {
					frontend.BindModifier = value.(string)
				}
				if value, ok := val[i]["certificates"]; ok {
					frontend.Certificates = value.([]string)
				}
				if value, ok := val[i]["misc_strs"]; ok {
					frontend.MiscStrs = value.([]string)
				}

				if value, ok := val[i]["redirect_to_https"]; ok {
					except := value.(map[string]interface{})
					if val, ok := except["except"]; ok {
						rth := dcos.EdgelbV2FrontendRedirectToHttps{}
						rth.Except = make([]dcos.EdgelbV2FrontendRedirectToHttpsExcept, 0)
						for _, vals := range val.(*schema.Set).List() {
							if e, ok := vals.(map[string]interface{}); ok {
								ex := dcos.EdgelbV2FrontendRedirectToHttpsExcept{}
								ex.Host = e["host"].(string)
								ex.PathBeg = e["path_beg"].(string)

								rth.Except = append(rth.Except, ex)
							}
						}

						if len(rth.Except) > 0 {
							frontend.RedirectToHttps = &rth
						}
					}
				}
				frontends = append(frontends, frontend)
			}

			edgelbV2Pool.Haproxy.Frontends = frontends
		}

		if v, ok := d.GetOk("haproxy.backends"); ok {
			backends := make([]dcos.EdgelbV2Backend, 0)
			for i := range v.(*schema.Set).List() {
				val := v.([]map[string]interface{})
				backend := dcos.EdgelbV2Backend{}

				if value, ok := val[i]["name"]; ok {
					backend.Name = value.(string)
				}

				if value, ok := val[i]["protocol"]; ok {
					switch value.(string) {
					case "TCP":
						backend.Protocol = dcos.EdgelbV2ProtocolTCP
					case "TLS":
						backend.Protocol = dcos.EdgelbV2ProtocolTLS
					case "HTTP":
						backend.Protocol = dcos.EdgelbV2ProtocolHTTP
					case "HTTPS":
						backend.Protocol = dcos.EdgelbV2ProtocolHTTPS
					default:
						return edgelbV2Pool, fmt.Errorf("Unknown protocol - %s", value.(string))
					}
				}

				if value, ok := val[i]["rewrite_http"]; ok {
					rhttp := value.(map[string]interface{})
					backend.RewriteHttp = dcos.EdgelbV2RewriteHttp{}

					if host, ok := rhttp["host"]; ok {
						backend.RewriteHttp.Host = host.(string)
					}

					if p, ok := rhttp["path"]; ok {
						path := p.(map[string]interface{})

						if f, ok := path["from_path"]; ok {
							backend.RewriteHttp.Path.FromPath = f.(string)
						}

						if f, ok := path["to_path"]; ok {
							backend.RewriteHttp.Path.ToPath = f.(string)
						}
					}

					if r, ok := rhttp["request"]; ok {
						request := r.(map[string]interface{})
						backend.RewriteHttp.Request = dcos.EdgelbV2RewriteHttpRequest{}

						if f, ok := request["forwardedfor"]; ok {
							backend.RewriteHttp.Request.Forwardfor = f.(bool)
						}
						if f, ok := request["x_forwarded_port"]; ok {
							backend.RewriteHttp.Request.XForwardedPort = f.(bool)
						}
						if f, ok := request["x_forwarded_proto_https_if_tls"]; ok {
							backend.RewriteHttp.Request.XForwardedProtoHttpsIfTls = f.(bool)
						}
						if f, ok := request["set_host_header"]; ok {
							backend.RewriteHttp.Request.SetHostHeader = f.(bool)
						}
						if f, ok := request["rewrite_path"]; ok {
							backend.RewriteHttp.Request.RewritePath = f.(bool)
						}
					}

					if r, ok := rhttp["response"]; ok {
						response := r.(map[string]interface{})
						backend.RewriteHttp.Response = dcos.EdgelbV2RewriteHttpResponse{}

						if f, ok := response["rewrite_location"]; ok {
							backend.RewriteHttp.Response.RewriteLocation = f.(bool)
						}
					}

					if s, ok := rhttp["sticky"]; ok {
						sticky := s.(map[string]interface{})
						backend.RewriteHttp.Sticky = dcos.EdgelbV2RewriteHttpSticky{}

						if f, ok := sticky["enabled"]; ok {
							backend.RewriteHttp.Sticky.Enabled = f.(bool)
						}
						if f, ok := sticky["custom_str"]; ok {
							backend.RewriteHttp.Sticky.CustomStr = f.(string)
						}
					}
				}

				if value, ok := val[i]["balance"]; ok {
					backend.Balance = value.(string)
				}

				if value, ok := val[i]["custom_check"]; ok {
					customCheck := value.(map[string]interface{})
					backend.CustomCheck = dcos.EdgelbV2BackendCustomCheck{}

					if val, ok := customCheck["httpchk"]; ok {
						backend.CustomCheck.Httpchk = val.(bool)
					}
					if val, ok := customCheck["httpchk_misc_str"]; ok {
						backend.CustomCheck.HttpchkMiscStr = val.(string)
					}
					if val, ok := customCheck["ssl_hello_chk"]; ok {
						backend.CustomCheck.SslHelloChk = val.(bool)
					}
					if val, ok := customCheck["misc_str"]; ok {
						backend.CustomCheck.MiscStr = val.(string)
					}
				}

				if value, ok := val[i]["misc_strs"]; ok {
					backend.MiscStrs = value.([]string)
				}

				if value, ok := val[i]["services"]; ok {
					backend.Services = make([]dcos.EdgelbV2Service, 0)
					for _, val := range value.(*schema.Set).List() {
						if s, ok := val.(map[string]interface{}); ok {
							service := dcos.EdgelbV2Service{}
							if marathon, ok := s["marathon"]; ok {
								m := marathon.(map[string]interface{})
								service.Marathon = dcos.EdgelbV2ServiceMarathon{}
								if v, ok := m["service_id"]; ok {
									service.Marathon.ServiceID = v.(string)
								}
								if v, ok := m["service_id_pattern"]; ok {
									service.Marathon.ServiceIDPattern = v.(string)
								}
								if v, ok := m["container_name"]; ok {
									service.Marathon.ContainerName = v.(string)
								}
								if v, ok := m["container_name_pattern"]; ok {
									service.Marathon.ContainerNamePattern = v.(string)
								}
							}

							if mesos, ok := s["mesos"]; ok {
								m := mesos.(map[string]interface{})
								service.Mesos = dcos.EdgelbV2ServiceMesos{}
								if v, ok := m["framework_name"]; ok {
									service.Mesos.FrameworkName = v.(string)
								}
								if v, ok := m["framework_name_pattern"]; ok {
									service.Mesos.FrameworkNamePattern = v.(string)
								}
								if v, ok := m["framework_id"]; ok {
									service.Mesos.FrameworkID = v.(string)
								}
								if v, ok := m["framework_id_pattern"]; ok {
									service.Mesos.FrameworkIDPattern = v.(string)
								}
								if v, ok := m["task_name"]; ok {
									service.Mesos.TaskName = v.(string)
								}
								if v, ok := m["task_name_pattern"]; ok {
									service.Mesos.TaskNamePattern = v.(string)
								}
								if v, ok := m["task_id"]; ok {
									service.Mesos.TaskID = v.(string)
								}
								if v, ok := m["task_id_pattern"]; ok {
									service.Mesos.TaskIDPattern = v.(string)
								}
							}

							if endpoint, ok := s["endpoint"]; ok {
								e := endpoint.(map[string]interface{})
								service.Endpoint = dcos.EdgelbV2Endpoint{}
								if v, ok := e["type"]; ok {
									service.Endpoint.Type = v.(string)
								}
								if v, ok := e["misc_str"]; ok {
									service.Endpoint.MiscStr = v.(string)
								}
								if check, ok := e["check"]; ok {
									c := check.(map[string]interface{})
									service.Endpoint.Check = dcos.EdgelbV2EndpointCheck{}

									if v, ok := c["enabled"]; ok {
										service.Endpoint.Check.Enabled = v.(bool)
									}

									if v, ok := c["custom_str"]; ok {
										service.Endpoint.Check.CustomStr = v.(string)
									}
								}
								if v, ok := e["address"]; ok {
									service.Endpoint.Address = v.(string)
								}
								if v, ok := e["port"]; ok {
									service.Endpoint.Port = v.(int32)
								}
								if v, ok := e["port_name"]; ok {
									service.Endpoint.PortName = v.(string)
								}
								if v, ok := e["all_ports"]; ok {
									service.Endpoint.AllPorts = v.(bool)
								}
							}
							backend.Services = append(backend.Services, service)
						}
					}

					// s := value.(map[string]interface{})
					// backend.Services = make([]dcos.EdgelbV2Service, 0)
					//
					// if val, ok := customCheck["httpchk"]; ok {
					// 	backend.CustomCheck.Httpchk = val.(bool)
					// }
					// if val, ok := customCheck["httpchk_misc_str"]; ok {
					// 	backend.CustomCheck.HttpchkMiscStr = val.(string)
					// }
					// if val, ok := customCheck["ssl_hello_chk"]; ok {
					// 	backend.CustomCheck.SslHelloChk = val.(bool)
					// }
					// if val, ok := customCheck["misc_str"]; ok {
					// 	backend.CustomCheck.MiscStr = val.(string)
					// }
				}

			}
			edgelbV2Pool.Haproxy.Backends = backends
		}

	}

	return edgelbV2Pool, nil
}

func resourceDcosEdgeLBV2PoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	edgelbV2Pool, err := edgelbV2PoolFromSchema(d)
	if err != nil {
		return err
	}

	client.Edgelb.V2CreatePool(ctx, edgelbV2Pool)

	return resourceDcosEdgeLBV2PoolRead(d, meta)
}

func resourceDcosEdgeLBV2PoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	poolName := d.Get("name").(string)

	pool, resp, err := client.Edgelb.V2GetPool(ctx, poolName)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.Set("pool_healthcheck_grace_period", pool.PoolHealthcheckGracePeriod)
	d.Set("pool_healthcheck_interval", pool.PoolHealthcheckInterval)
	d.Set("pool_healthcheck_max_fail", pool.PoolHealthcheckMaxFail)
	d.Set("pool_healthcheck_timeout", pool.PoolHealthcheckTimeout)

	d.Set("namespace", pool.Namespace)
	d.Set("role", pool.Role)
	d.Set("principal", pool.Principal)
	d.Set("secret_name", pool.SecretName)

	d.Set("cpus", pool.Cpus)
	d.Set("mem", pool.Mem)
	d.Set("disk", pool.Disk)
	d.Set("count", pool.Count)

	d.Set("constraints", pool.Constraints)
	d.Set("ports", pool.Ports)

	if len(pool.Secrets) > 0 {
		secrets := make([]map[string]interface{}, 0)

		for _, secret := range pool.Secrets {
			secrets = append(secrets, map[string]interface{}{
				"secret": secret.Secret,
				"file":   secret.File,
			})
		}

		d.Set("secrets", secrets)
	}

	if len(pool.EnvironmentVariables) > 0 {
		d.Set("environment_variables", pool.EnvironmentVariables)
	}

	d.Set("auto_certificate", pool.AutoCertificate)

	if len(pool.VirtualNetworks) > 0 {
		virtualNetworks := make([]map[string]interface{}, 0)

		for _, network := range pool.VirtualNetworks {
			virtualNetworks = append(virtualNetworks, map[string]interface{}{
				"name":   network.Name,
				"labels": network.Labels,
			})
		}

		d.Set("virtual_networks", virtualNetworks)
	}

	if len(pool.Haproxy.Frontends) > 0 {
		frontends := make([]map[string]interface{}, 0)

		for _, frontend := range pool.Haproxy.Frontends {
			f := make(map[string]interface{})

			f["name"] = frontend.Name
			f["bind_address"] = frontend.BindAddress
			f["bind_port"] = frontend.BindPort
			f["bind_modifier"] = frontend.BindModifier
			f["certificates"] = frontend.Certificates

			if frontend.RedirectToHttps != nil && len(frontend.RedirectToHttps.Except) > 0 {
				except := make([]map[string]interface{}, 0)

				for _, e := range frontend.RedirectToHttps.Except {
					except = append(except, map[string]interface{}{
						"host":     e.Host,
						"path_beg": e.PathBeg,
					})
				}
				f["redirect_to_https"] = map[string]interface{}{
					"except": except,
				}
			}

			if len(frontend.MiscStrs) > 0 {
				f["misc_strs"] = frontend.MiscStrs
			}

			frontends = append(frontends, f)
		}

		d.Set("haproxy.frontends", frontends)
	}

	if len(pool.Haproxy.Backends) > 0 {
		backends := make([]map[string]interface{}, 0)
		for _, backend := range pool.Haproxy.Backends {
			b := make(map[string]interface{})

			b["name"] = backend.Name
			b["protocol"] = backend.Protocol
			b["rewrite_http"] = map[string]interface{}{
				"host": backend.RewriteHttp.Host,
				"path": map[string]interface{}{
					"from_path": backend.RewriteHttp.Path.FromPath,
					"to_path":   backend.RewriteHttp.Path.ToPath,
				},
				"request": map[string]interface{}{
					"forwardfor":                     backend.RewriteHttp.Request.Forwardfor,
					"x_forwarded_port":               backend.RewriteHttp.Request.XForwardedPort,
					"x_forwarded_proto_https_if_tls": backend.RewriteHttp.Request.XForwardedProtoHttpsIfTls,
					"set_host_header":                backend.RewriteHttp.Request.SetHostHeader,
					"rewrite_path":                   backend.RewriteHttp.Request.RewritePath,
				},
				"response": map[string]interface{}{
					"rewrite_location": backend.RewriteHttp.Response.RewriteLocation,
				},
				"sticky": map[string]interface{}{
					"enabled":    backend.RewriteHttp.Sticky.Enabled,
					"custom_str": backend.RewriteHttp.Sticky.CustomStr,
				},
			}
			b["balance"] = backend.Balance
			b["custom_check"] = map[string]interface{}{
				"httpchk":          backend.CustomCheck.Httpchk,
				"httpchk_misc_str": backend.CustomCheck.HttpchkMiscStr,
				"ssl_hello_chk":    backend.CustomCheck.SslHelloChk,
				"misc_str":         backend.CustomCheck.MiscStr,
			}
			b["misc_strs"] = backend.MiscStrs

			if len(backend.Services) > 0 {
				services := make([]map[string]interface{}, 0)
				for _, service := range backend.Services {
					s := make(map[string]interface{})
					s["marathon"] = map[string]interface{}{
						"service_id":             service.Marathon.ServiceID,
						"service_id_pattern":     service.Marathon.ServiceIDPattern,
						"container_name":         service.Marathon.ContainerName,
						"container_name_pattern": service.Marathon.ContainerNamePattern,
					}
					s["mesos"] = map[string]interface{}{
						"framework_name":         service.Mesos.FrameworkName,
						"framework_name_pattern": service.Mesos.FrameworkNamePattern,
						"framework_id":           service.Mesos.FrameworkID,
						"framework_id_pattern":   service.Mesos.FrameworkIDPattern,
						"task_name":              service.Mesos.TaskName,
						"task_name_pattern":      service.Mesos.TaskNamePattern,
						"task_id":                service.Mesos.TaskID,
						"task_id_pattern":        service.Mesos.TaskIDPattern,
					}
					s["endpoint"] = map[string]interface{}{
						"type": service.Endpoint.Type,
						"check": map[string]interface{}{
							"enabled":    service.Endpoint.Check.Enabled,
							"custom_str": service.Endpoint.Check.CustomStr,
						},
						"address":   service.Endpoint.Address,
						"port":      service.Endpoint.Port,
						"port_name": service.Endpoint.PortName,
						"all_ports": service.Endpoint.AllPorts,
					}
					services = append(services, s)
				}

				b["services"] = services
			}
		}

		d.Set("backends", backends)
	}

	d.SetId(poolName)

	return nil
}

func resourceDcosEdgeLBV2PoolUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()
	//
	// var iamsamlProviderConfig dcos.IamsamlProviderConfig
	//
	// providerId := d.Get("provider_id").(string)
	// idpMetadata := d.Get("idp_metadata").(string)
	// spBaseURL := d.Get("base_url").(string)
	//
	// if description, ok := d.GetOk("description"); ok {
	// 	iamsamlProviderConfig.Description = description.(string)
	// }
	//
	// iamsamlProviderConfig.IdpMetadata = idpMetadata
	// iamsamlProviderConfig.SpBaseUrl = spBaseURL
	//
	// resp, err := client.IAM.UpdateSAMLProvider(ctx, providerId, iamsamlProviderConfig)
	//
	// log.Printf("[TRACE] IAM.UpdateSAMLProvider - %v", resp)
	//
	// if err != nil {
	// 	return err
	// }

	return resourceDcosEdgeLBV2PoolRead(d, meta)
}

func resourceDcosEdgeLBV2PoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	poolName := d.Get("name").(string)

	resp, err := client.Edgelb.V2DeletePool(ctx, poolName)

	log.Printf("[TRACE] Edgelb.V2DeletePool - %v", resp)

	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
