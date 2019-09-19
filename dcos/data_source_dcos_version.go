package dcos

import (
	"github.com/dcos/client-go/dcos"
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
	client := meta.(*dcos.APIClient)
	ver, err := util.DCOSGetVersion(client)
	if err != nil {
		return err
	}

	d.Set("version", ver.Version)
	d.Set("dcos_variant", ver.DcosVariant)
	d.Set("dcos_image_commit", ver.DcosImageCommit)
	d.Set("bootstrap_id", ver.BootstrapId)

	d.SetId(ver.Version)

	return nil
}
