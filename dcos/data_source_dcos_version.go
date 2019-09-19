package dcos

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

func dataSourceDcosVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosVersionRead,
		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDcosVersionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	value, err := util.DCOSGetVersion(client)
	if err != nil {
		return err
	}

	d.Set("version", value)
	d.SetId(value)

	return nil
}
