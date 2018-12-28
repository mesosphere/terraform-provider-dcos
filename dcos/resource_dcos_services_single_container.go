package dcos

import (
	"time"

	marathon "github.com/gambol99/go-marathon"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mesosphere/dcos-go/dcos"
)

func resourceDcosServicesSingleContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosServicesSingleContainerCreate,
		Read:   resourceDcosServicesSingleContainerRead,
		Update: resourceDcosServicesSingleContainerUpdate,
		Delete: resourceDcosServicesSingleContainerDelete,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instances": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cmd": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cpus": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"disk": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"mem": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
		},
	}
}

func resourceDcosServicesSingleContainerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.Client)

	application := marathon.Application{
		ID:        d.Get("name").(string),
		Cmd:       d.Get("cmd").(*string),
		Instances: d.Get("instances").(*int),
		CPUs:      d.Get("cpus").(float64),
		Disk:      d.Get("disk").(*float64),
		Mem:       d.Get("mem").(*float64),
	}

	app, err := client.Marathon.MarathonClient.CreateApplication(&application)

	if err != nil {
		return err
	}

	d.SetId(app.ID)

	return resourceDcosServicesSingleContainerRead(d, meta)
}

func resourceDcosServicesSingleContainerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.Client)

	app, err := client.Marathon.MarathonClient.Application(d.Id())

	d.Set("cmd", app.Cmd)
	d.Set("instances", app.Instances)
	d.Set("cpus", app.CPUs)
	d.Set("disk", app.Disk)
	d.Set("mem", app.Mem)

	return err
}

func resourceDcosServicesSingleContainerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.Client)

	app, err := client.Marathon.MarathonClient.Application(d.Id())
	if err != nil {
		return err
	}

	d.Set("cmd", app.Cmd)

	if d.HasChange("cmd") {
		app.Cmd = d.Get("cmd").(*string)
	}
	if d.HasChange("instances") {
		app.Instances = d.Get("instances").(*int)
	}
	if d.HasChange("cpus") {
		app.CPUs = d.Get("cpus").(float64)
	}
	d.Set("disk", app.Disk)
	if d.HasChange("disk") {
		app.Disk = d.Get("disk").(*float64)
	}

	d.Set("mem", app.Mem)
	if d.HasChange("mem") {
		app.Mem = d.Get("mem").(*float64)
	}

	_, err = client.Marathon.MarathonClient.UpdateApplication(app, false)

	if err != nil {
		return err
	}

	return resourceDcosServicesSingleContainerRead(d, meta)
}

func resourceDcosServicesSingleContainerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.Client)

	_, err := client.Marathon.MarathonClient.DeleteApplication(d.Id(), false)

	return err
}
