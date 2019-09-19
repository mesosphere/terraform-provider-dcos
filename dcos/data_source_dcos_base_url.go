package dcos

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDcosBaseURL() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosBaseURLRead,
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDcosBaseURLRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	//ctx := context.TODO()

	dcosConfig := client.CurrentDCOSConfig()
	url := dcosConfig.URL()
	d.Set("url", url)
	d.SetId(url)

	return nil
}
