package dcos

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDcosCLICommand() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosCLICommandRead,
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDcosCLICommandRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	//ctx := context.TODO()

	cli, err := util.CreateCliWrapper(".terraform/dcos/sandbox", client, d.Get("cli_version").(string))
	if err != nil {
		return nil, fmt.Errorf("Unable to create cli wrapper: %s", err.Error())
	}

	dcosConfig := client.CurrentDCOSConfig()
	url := dcosConfig.URL()
	d.Set("url", url)
	d.SetId(url)

	return nil
}
