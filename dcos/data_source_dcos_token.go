package dcos

import (
	// "github.com/dcos/client-go/dcos"

	"github.com/dcos/client-go/dcos"
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
	client := meta.(*dcos.APIClient)
	//ctx := context.TODO()

	dcosConfig := client.CurrentDCOSConfig()
	token := dcosConfig.ACSToken()
	d.Set("token", token)
	d.SetId(token)

	return nil
}
