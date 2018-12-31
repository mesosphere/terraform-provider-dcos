package dcos

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mesosphere/dcos-go/dcos"
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
		},
		ResourcesMap: map[string]*schema.Resource{
			"dcos_services_single_container": resourceDcosServicesSingleContainer(),
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
