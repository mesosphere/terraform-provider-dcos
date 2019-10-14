package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
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
				Computed:    true,
				ForceNew:    false,
				Description: "Pool tasks healthcheck grace period (in seconds)",
			},
			"pool_healthcheck_interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    false,
				Description: "Pool tasks healthcheck interval (in seconds)",
			},
			"pool_healthcheck_max_fail": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    false,
				Description: "Pool tasks healthcheck maximum number of consecutive failures before declaring as unhealthy",
			},
			"pool_healthcheck_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    false,
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
				Computed:    true,
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
				Computed:    true,
				ForceNew:    false,
				Description: "Disk size (in MB)",
			},
			"pool_count": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Description: "Number of load balancer instances in the pool",
			},
			"constraints": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    false,
				Description: "Marathon style constraints for load balancer instance placement",
			},
			"ports": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
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

			"haproxy_frontends": {
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
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"redirect_to_https_except": {
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
						"misc_strs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional template lines inserted before use_backend",
						},
						"protocol": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Protocol",
						},

						"linked_backend_default_backend": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "This is default backend that is routed to if none of the other filters are matched.",
						},
						"linked_backend_map": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    false,
							Description: "This is an optional field that specifies a mapping to various backends. These rules are applied in order.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"backend": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"host_eq": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"host_reg": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"path_beg": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"path_end": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"path_reg": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"haproxy_backends": {
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
						"rewrite_http_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"rewrite_http_from_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"rewrite_http_to_path": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"rewrite_http_request_forwardfor": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Set X-Forwarded-For",
						},
						"rewrite_http_request_x_forwarded_port": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Set X-Forwarded-Port",
						},
						"rewrite_http_request_x_forwarded_proto_https_if_tls": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Set X-Forwarded-Port HTTPS if TLS",
						},
						"rewrite_http_request_set_host_header": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Set Host header",
						},
						"rewrite_http_request_rewrite_path": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Rewrite Path",
						},

						"rewrite_http_response_rewrite_location": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"rewrite_http_sticky_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"rewrite_http_sticky_custom_str": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"balance": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Load balancing strategy. e.g. roundrobin, leastconn, etc.",
						},
						"custom_check_httpchk": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"custom_check_httpchk_misc_str": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"custom_check_ssl_hello_chk": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"custom_check_misc_str": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the virtual network to join.",
						},
						"misc_strs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Additional template lines inserted before servers",
						},
						"services": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"marathon_service_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Marathon pod or application ID",
									},
									"marathon_service_id_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},
									"marathon_container_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Marathon pod container name, optional unless using Marathon pods",
									},
									"marathon_container_name_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},

									"mesos_framework_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Mesos framework name",
									},
									"mesos_framework_name_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},
									"mesos_framework_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Mesos framework ID",
									},
									"mesos_framework_id_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},
									"mesos_task_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Mesos task name",
									},
									"mesos_task_name_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},
									"mesos_task_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Mesos task ID",
									},
									"mesos_task_id_pattern": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},

									"endpoint_type": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "Mesos framework name",
									},
									"endpoint_misc_str": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Append arbitrary string to add to the end of the \"server\" directive",
									},
									"endpoint_check_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										// Description: "Mesos framework name",
									},
									"endpoint_check_custom_str": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "Append arbitrary string to add to the end of the \"server\" directive",
									},

									"endpoint_address": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Server address override, can be used to specify a cluster internal address such as a VIP",
									},
									"endpoint_port": {
										Type:     schema.TypeInt,
										Optional: true,
										// Description: "Mesos task name",
									},
									"endpoint_port_name": {
										Type:     schema.TypeString,
										Optional: true,
										// Description: "The name of the virtual network to join.",
									},
									"endpoint_all_ports": {
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
	}
}

func edgelbV2PoolFromSchema(d *schema.ResourceData) (dcos.EdgelbV2Pool, error) {
	edgelbV2Pool := dcos.EdgelbV2Pool{}

	poolName := d.Get("name").(string)

	edgelbV2Pool.Name = poolName

	if v, ok := d.GetOk("pool_healthcheck_grace_period"); ok {
		edgelbV2Pool.PoolHealthcheckGracePeriod = int32(v.(int))
	}
	if v, ok := d.GetOk("pool_healthcheck_interval"); ok {
		edgelbV2Pool.PoolHealthcheckInterval = int32(v.(int))
	}
	if v, ok := d.GetOk("pool_healthcheck_max_fail"); ok {
		edgelbV2Pool.PoolHealthcheckMaxFail = int32(v.(int))
	}
	if v, ok := d.GetOk("pool_healthcheck_timeout"); ok {
		edgelbV2Pool.PoolHealthcheckTimeout = int32(v.(int))
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
	if v, ok := d.GetOk("pool_count"); ok {
		edgelbV2Pool.Count = int32(v.(int))
	}
	if v, ok := d.GetOk("constraints"); ok {
		edgelbV2Pool.Constraints = v.(string)
	}
	if v, ok := d.GetOk("ports"); ok {
		edgelbV2Pool.Ports, _ = util.InterfaceSliceInt32(v.([]interface{}))
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

	if v, ok := d.GetOk("haproxy_frontends"); ok {
		frontends := make([]dcos.EdgelbV2Frontend, 0)
		for _, va := range v.(*schema.Set).List() {
			val := va.(map[string]interface{})
			frontend := dcos.EdgelbV2Frontend{}

			if value, ok := val["name"]; ok {
				frontend.Name = value.(string)
			}
			if value, ok := val["bind_port"]; ok {
				frontend.BindPort = int32(value.(int))
			}
			if value, ok := val["bind_modifier"]; ok {
				frontend.BindModifier = value.(string)
			}
			if value, ok := val["certificates"]; ok {
				frontend.Certificates, _ = util.InterfaceSliceString(value.([]interface{}))
			}
			if value, ok := val["misc_strs"]; ok {
				frontend.MiscStrs, _ = util.InterfaceSliceString(value.([]interface{}))
			}

			if value, ok := val["protocol"]; ok {
				switch value.(string) {
				case "TCP":
					frontend.Protocol = dcos.EdgelbV2ProtocolTCP
				case "TLS":
					frontend.Protocol = dcos.EdgelbV2ProtocolTLS
				case "HTTP":
					frontend.Protocol = dcos.EdgelbV2ProtocolHTTP
				case "HTTPS":
					frontend.Protocol = dcos.EdgelbV2ProtocolHTTPS
				default:
					return edgelbV2Pool, fmt.Errorf("Unknown protocol - %s", value.(string))
				}
			}

			linkedBackend := dcos.EdgelbV2FrontendLinkBackend{}
			if value, ok := val["linked_backend_default_backend"]; ok {
				linkedBackend.DefaultBackend = value.(string)
			}
			if value, ok := val["linked_backend_map"]; ok {
				linkedBackend.Map = make([]dcos.EdgelbV2FrontendLinkBackendMap, 0)
				for _, vals := range value.(*schema.Set).List() {
					if m, ok := vals.(map[string]interface{}); ok {
						ma := dcos.EdgelbV2FrontendLinkBackendMap{}
						ma.Backend = m["backend"].(string)
						ma.HostEq = m["host_eq"].(string)
						ma.HostReg = m["host_reg"].(string)
						ma.PathBeg = m["path_beg"].(string)
						ma.PathEnd = m["path_end"].(string)
						ma.PathReg = m["path_reg"].(string)

						linkedBackend.Map = append(linkedBackend.Map, ma)
					}
				}
			}
			frontend.LinkBackend = linkedBackend

			if val, ok := val["redirect_to_https_except"]; ok {
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
			frontends = append(frontends, frontend)
		}

		edgelbV2Pool.Haproxy.Frontends = frontends
	}

	if v, ok := d.GetOk("haproxy_backends"); ok {
		backends := make([]dcos.EdgelbV2Backend, 0)
		for _, va := range v.(*schema.Set).List() {
			val := va.(map[string]interface{})
			backend := dcos.EdgelbV2Backend{}

			if value, ok := val["name"]; ok {
				backend.Name = value.(string)
			}

			if value, ok := val["protocol"]; ok {
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

			backend.RewriteHttp = dcos.EdgelbV2RewriteHttp{}
			backend.RewriteHttp.Path = dcos.EdgelbV2RewriteHttpPath{}
			backend.RewriteHttp.Request = dcos.EdgelbV2RewriteHttpRequest{}
			backend.RewriteHttp.Response = dcos.EdgelbV2RewriteHttpResponse{}
			backend.RewriteHttp.Sticky = dcos.EdgelbV2RewriteHttpSticky{}

			if value, ok := val["rewrite_http_host"]; ok {
				backend.RewriteHttp.Host = value.(string)
			}
			if value, ok := val["rewrite_http_from_path"]; ok {
				backend.RewriteHttp.Path.FromPath = value.(string)
			}

			if value, ok := val["rewrite_http_to_path"]; ok {
				backend.RewriteHttp.Path.ToPath = value.(string)
			}

			if value, ok := val["rewrite_http_requst_forwardedfor"]; ok {
				backend.RewriteHttp.Request.Forwardfor = value.(bool)
			}
			if value, ok := val["rewrite_http_requst_x_forwarded_port"]; ok {
				backend.RewriteHttp.Request.XForwardedPort = value.(bool)
			}
			if value, ok := val["rewrite_http_requst_x_forwarded_proto_https_if_tls"]; ok {
				backend.RewriteHttp.Request.XForwardedProtoHttpsIfTls = value.(bool)
			}
			if value, ok := val["rewrite_http_requst_set_host_header"]; ok {
				backend.RewriteHttp.Request.SetHostHeader = value.(bool)
			}
			if value, ok := val["rewrite_http_requst_rewrite_path"]; ok {
				backend.RewriteHttp.Request.RewritePath = value.(bool)
			}

			if value, ok := val["rewrite_http_response_rewrite_location"]; ok {
				backend.RewriteHttp.Response.RewriteLocation = value.(bool)
			}

			if value, ok := val["rewrite_http_sticky_enabled"]; ok {
				backend.RewriteHttp.Sticky.Enabled = value.(bool)
			}

			if value, ok := val["rewrite_http_sticky_custom_str"]; ok {
				backend.RewriteHttp.Sticky.CustomStr = value.(string)
			}

			if value, ok := val["balance"]; ok {
				backend.Balance = value.(string)
			}

			backend.CustomCheck = dcos.EdgelbV2BackendCustomCheck{}
			if value, ok := val["custom_check_httpchk"]; ok {
				backend.CustomCheck.Httpchk = value.(bool)
			}
			if value, ok := val["custom_check_httpchk_misc_str"]; ok {
				backend.CustomCheck.HttpchkMiscStr = value.(string)
			}
			if value, ok := val["custom_check_ssl_hello_chk"]; ok {
				backend.CustomCheck.SslHelloChk = value.(bool)
			}
			if value, ok := val["custom_check_misc_str"]; ok {
				backend.CustomCheck.MiscStr = value.(string)
			}

			if value, ok := val["misc_strs"]; ok {
				backend.MiscStrs, _ = util.InterfaceSliceString(value.([]interface{}))
			}

			if value, ok := val["services"]; ok {
				backend.Services = make([]dcos.EdgelbV2Service, 0)
				for _, val := range value.(*schema.Set).List() {
					if s, ok := val.(map[string]interface{}); ok {
						service := dcos.EdgelbV2Service{}
						service.Marathon = dcos.EdgelbV2ServiceMarathon{}
						if v, ok := s["marathon_service_id"]; ok {
							service.Marathon.ServiceID = v.(string)
						}
						if v, ok := s["marathon_service_id_pattern"]; ok {
							service.Marathon.ServiceIDPattern = v.(string)
						}
						if v, ok := s["marathon_container_name"]; ok {
							service.Marathon.ContainerName = v.(string)
						}
						if v, ok := s["marathon_container_name_pattern"]; ok {
							service.Marathon.ContainerNamePattern = v.(string)
						}

						service.Mesos = dcos.EdgelbV2ServiceMesos{}
						if v, ok := s["mesos_framework_name"]; ok {
							service.Mesos.FrameworkName = v.(string)
						}
						if v, ok := s["mesos_framework_name_pattern"]; ok {
							service.Mesos.FrameworkNamePattern = v.(string)
						}
						if v, ok := s["mesos_framework_id"]; ok {
							service.Mesos.FrameworkID = v.(string)
						}
						if v, ok := s["mesos_framework_id_pattern"]; ok {
							service.Mesos.FrameworkIDPattern = v.(string)
						}
						if v, ok := s["mesos_task_name"]; ok {
							service.Mesos.TaskName = v.(string)
						}
						if v, ok := s["mesos_task_name_pattern"]; ok {
							service.Mesos.TaskNamePattern = v.(string)
						}
						if v, ok := s["mesos_task_id"]; ok {
							service.Mesos.TaskID = v.(string)
						}
						if v, ok := s["mesos_task_id_pattern"]; ok {
							service.Mesos.TaskIDPattern = v.(string)
						}

						service.Endpoint = dcos.EdgelbV2Endpoint{}
						if v, ok := s["endpoint_type"]; ok {
							service.Endpoint.Type = v.(string)
						}
						if v, ok := s["endpoint_misc_str"]; ok {
							service.Endpoint.MiscStr = v.(string)
						}
						service.Endpoint.Check = dcos.EdgelbV2EndpointCheck{}

						if v, ok := s["endpoint_check_enabled"]; ok {
							service.Endpoint.Check.Enabled = v.(bool)
						}
						if v, ok := s["endpoint_check_custom_str"]; ok {
							service.Endpoint.Check.CustomStr = v.(string)
						}

						if v, ok := s["endpoint_address"]; ok {
							service.Endpoint.Address = v.(string)
						}
						if v, ok := s["endpoint_port"]; ok {
							service.Endpoint.Port = int32(v.(int))
						}
						if v, ok := s["endpoint_port_name"]; ok {
							service.Endpoint.PortName = v.(string)
						}
						if v, ok := s["endpoint_all_ports"]; ok {
							service.Endpoint.AllPorts = v.(bool)
						}

						backend.Services = append(backend.Services, service)
					}
				}
			}
			backends = append(backends, backend)
		}
		edgelbV2Pool.Haproxy.Backends = backends
	}

	log.Printf("[TRACE] edgelbV2PoolFromSchema - calculated EdgelbV2Pool %+v", edgelbV2Pool)

	return edgelbV2Pool, nil
}

func pingEdgeLB(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	p, resp, err := client.Edgelb.Ping(ctx)
	log.Printf("[TRACE] Edgelb.Ping - p: %s, resp: %v", p, resp)

	return err
}

func pingEdgeLBRetryFunc(d *schema.ResourceData, meta interface{}) func() *resource.RetryError {
	return func() *resource.RetryError {
		err := pingEdgeLB(d, meta)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	}
}

func resourceDcosEdgeLBV2PoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	if err := resource.Retry(d.Timeout(schema.TimeoutCreate), pingEdgeLBRetryFunc(d, meta)); err != nil {
		return err
	}

	edgelbV2Pool, err := edgelbV2PoolFromSchema(d)
	if err != nil {
		return err
	}

	_, resp, err := client.Edgelb.V2CreatePool(ctx, edgelbV2Pool)

	log.Printf("[TRACE] Edgelb.V2CreatePool - %v", resp)

	if err != nil {
		if apiError, ok := err.(dcos.GenericOpenAPIError); ok {

			log.Printf("[ERROR] Edgelb.V2CreatePool - ==========BODY=======%s==========BODY=======", string(apiError.Body()))
		}

		if resp != nil && resp.StatusCode != http.StatusInternalServerError {
			// DCOS-59682 we try read if we face an internal server errror
			return err
		}
	}

	return resourceDcosEdgeLBV2PoolRead(d, meta)
}

func resourceDcosEdgeLBV2PoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	var poolName string
	if d.Id() == "" {
		poolName = d.Get("name").(string)
	} else {
		poolName = d.Id()
	}

	pool, resp, err := client.Edgelb.V2GetPool(ctx, poolName)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	log.Printf("[TRACE] Edgelb.V2DeletePool - %v", pool)

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
	d.Set("pool_count", pool.Count)

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
				f["redirect_to_https_except"] = except
			}

			if len(frontend.MiscStrs) > 0 {
				f["misc_strs"] = frontend.MiscStrs
			}

			f["linked_backend_default_backend"] = frontend.LinkBackend.DefaultBackend

			lmaps := make([]map[string]interface{}, 0)
			if len(frontend.LinkBackend.Map) > 0 {
				for _, m := range frontend.LinkBackend.Map {
					ma := map[string]interface{}{
						"backend":  m.Backend,
						"host_eq":  m.HostEq,
						"host_reg": m.HostReg,
						"path_beg": m.PathBeg,
						"path_end": m.PathEnd,
						"path_reg": m.PathReg,
					}
					lmaps = append(lmaps, ma)
				}
			}

			f["linked_backend_map"] = lmaps

			frontends = append(frontends, f)
		}

		d.Set("haproxy_frontends", frontends)
	}

	if len(pool.Haproxy.Backends) > 0 {
		backends := make([]map[string]interface{}, 0)
		for _, backend := range pool.Haproxy.Backends {
			b := make(map[string]interface{})

			b["name"] = backend.Name
			b["protocol"] = backend.Protocol
			b["rewrite_http_host"] = backend.RewriteHttp.Host
			b["rewrite_http_from_path"] = backend.RewriteHttp.Path.FromPath
			b["rewrite_http_to_path"] = backend.RewriteHttp.Path.ToPath
			b["rewrite_http_request_forwardfor"] = backend.RewriteHttp.Request.Forwardfor
			b["rewrite_http_request_x_forwarded_port"] = backend.RewriteHttp.Request.XForwardedPort
			b["rewrite_http_request_x_forwarded_proto_https_if_tls"] = backend.RewriteHttp.Request.XForwardedProtoHttpsIfTls
			b["rewrite_http_request_set_host_header"] = backend.RewriteHttp.Request.SetHostHeader
			b["rewrite_http_request_rewrite_path"] = backend.RewriteHttp.Request.RewritePath
			b["rewrite_http_response_rewrite_location"] = backend.RewriteHttp.Response.RewriteLocation
			b["rewrite_http_sticky_enabled"] = backend.RewriteHttp.Sticky.Enabled
			b["rewrite_http_sticky_custom_str"] = backend.RewriteHttp.Sticky.CustomStr
			b["balance"] = backend.Balance
			b["custom_check_httpchk"] = backend.CustomCheck.Httpchk
			b["custom_check_httpchk_misc_str"] = backend.CustomCheck.HttpchkMiscStr
			b["custom_check_ssl_hello_chk"] = backend.CustomCheck.SslHelloChk
			b["custom_check_misc_str"] = backend.CustomCheck.MiscStr
			b["misc_strs"] = backend.MiscStrs

			if len(backend.Services) > 0 {
				services := make([]map[string]interface{}, 0)
				for _, service := range backend.Services {
					s := make(map[string]interface{})
					s["marathon_service_id"] = service.Marathon.ServiceID
					s["marathon_service_id_pattern"] = service.Marathon.ServiceIDPattern
					s["marathon_container_name"] = service.Marathon.ContainerName
					s["marathon_container_name_pattern"] = service.Marathon.ContainerNamePattern
					s["mesos_framework_name"] = service.Mesos.FrameworkName
					s["mesos_framework_name_pattern"] = service.Mesos.FrameworkNamePattern
					s["mesos_framework_id"] = service.Mesos.FrameworkID
					s["mesos_framework_id_pattern"] = service.Mesos.FrameworkIDPattern
					s["mesos_task_name"] = service.Mesos.TaskName
					s["mesos_task_name_pattern"] = service.Mesos.TaskNamePattern
					s["mesos_task_id"] = service.Mesos.TaskID
					s["mesos_task_id_pattern"] = service.Mesos.TaskIDPattern
					s["endpoint_type"] = service.Endpoint.Type
					s["endpoint_check_enabled"] = service.Endpoint.Check.Enabled
					s["endpoint_check_custom_str"] = service.Endpoint.Check.CustomStr
					s["address"] = service.Endpoint.Address
					s["port"] = service.Endpoint.Port
					s["port_name"] = service.Endpoint.PortName
					s["all_ports"] = service.Endpoint.AllPorts
					services = append(services, s)
				}

				b["services"] = services
			}
		}

		d.Set("haproxy_backends", backends)
	}

	d.SetId(poolName)

	return nil
}

func resourceDcosEdgeLBV2PoolUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	poolName := d.Get("name").(string)

	if err := pingEdgeLB(d, meta); err != nil {
		return err
	}

	edgelbV2Pool, err := edgelbV2PoolFromSchema(d)
	if err != nil {
		return err
	}

	_, resp, err := client.Edgelb.V2UpdatePool(ctx, poolName, edgelbV2Pool)

	log.Printf("[TRACE] Edgelb.V2UpdatePool - %v", resp)

	if err != nil {
		if apiError, ok := err.(dcos.GenericOpenAPIError); ok {

			log.Printf("[ERROR] Edgelb.V2UpdatePool - ==========BODY=======%s==========BODY=======", string(apiError.Body()))
		}

		if resp != nil && resp.StatusCode != http.StatusInternalServerError {
			// DCOS-59682 we try read if we face an internal server errror
			return err
		}
	}

	return resourceDcosEdgeLBV2PoolRead(d, meta)
}

func resourceDcosEdgeLBV2PoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	poolName := d.Get("name").(string)

	resp, err := client.Edgelb.V2DeletePool(ctx, poolName)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	log.Printf("[TRACE] Edgelb.V2DeletePool - %v", resp)

	if err != nil {
		if apiError, ok := err.(dcos.GenericOpenAPIError); ok {
			log.Printf("[ERROR] Edgelb.V2DeletePool - ==========BODY=======%s==========BODY=======", string(apiError.Body()))
		}
		return err
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, resp, err := client.Edgelb.V2GetPool(ctx, poolName)

		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		if err != nil {
			return resource.RetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("Pool %s still exists. Deleting... ", poolName))
	})
}
