package dcos

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mesosphere/dcos-go/dcos"
)

func dataSourceDcosService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosServiceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instances": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cmd": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpus": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"disk": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"mem": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func dataSourceDcosServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.Client)

	app, err := client.Marathon.MarathonClient.Application(d.Get("name").(string))

	if err != nil {
		return err
	}

	d.SetId(app.ID)

	d.Set("cmd", app.Cmd)
	d.Set("instances", app.Instances)
	d.Set("cpus", app.CPUs)
	d.Set("disk", app.Disk)
	d.Set("mem", app.Mem)

	return err
}
