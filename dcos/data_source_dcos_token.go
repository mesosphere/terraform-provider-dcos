package dcos

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDcosToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosTokenRead,
		Schema: map[string]*schema.Schema{
			"token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDcosTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	//ctx := context.TODO()

	dcosConfig := client.CurrentDCOSConfig()
	token := dcosConfig.ACSToken()
	d.Set("token", token)
	d.SetId(token)

	return nil
}
