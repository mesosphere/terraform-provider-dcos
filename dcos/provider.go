package dcos

import (
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"dcos_acs_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The DC/OS access token",
			},
			"ssl_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     "",
				Description: "Verify SSL connection",
			},
			"dcos_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "URL of DC/OS to use",
			},
			"cluster": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Clustername to use",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "User name logging into the cluster",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Password to login with",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"dcos_services_single_container": resourceDcosServicesSingleContainer(),
			"dcos_job":                       resourceDcosJob(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dcos_service": dataSourceDcosService(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	dcos, err := dcos.NewClient()

	return dcos, err
}
