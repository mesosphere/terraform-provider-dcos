package dcos

import (
	// "context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosCLI() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosCLICreate,
		Read:   resourceDcosCLIRead,
		Update: resourceDcosCLIUpdate,
		Delete: resourceDcosCLIDelete,

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the resource",
			},
			"package": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The DC/OS package from which to fetch the terminal",
			},

			"config": {
				Type:        schema.TypeString,
				Required:    true,
				Computed:    true,
				Description: "The configuration JSON to carry along",
			},

			"cmd_create": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The command to invoke for creating the resource",
			},
			"cmd_delete": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The command to invoke for deleting the resource",
			},
			"cmd_read": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The command to invoke for reading the resource",
			},
			"cmd_update": {
				Type:        schema.TypeString,
				Required:    false,
				ForceNew:    true,
				Default:     "",
				Description: "The command to invoke for updating the resource",
			},
		},
	}
}

func resourceDcosCLICreate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*ProviderState).Client
	// ctx := context.TODO()

	w := meta.(*ProviderState).CliWrapper
	_, err := w.ForPackage(d.Get("package").(string), d.Get("id").(string))
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	return nil
}

func resourceDcosCLIRead(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*ProviderState).Client
	// ctx := context.TODO()

	w := meta.(*ProviderState).CliWrapper
	_, err := w.ForPackage(d.Get("package").(string), d.Get("id").(string))
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	return nil
}

func resourceDcosCLIUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*ProviderState).Client
	// ctx := context.TODO()

	w := meta.(*ProviderState).CliWrapper
	_, err := w.ForPackage(d.Get("package").(string), d.Get("id").(string))
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	return nil
}

func resourceDcosCLIDelete(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*ProviderState).Client
	// ctx := context.TODO()

	w := meta.(*ProviderState).CliWrapper
	_, err := w.ForPackage(d.Get("package").(string), d.Get("id").(string))
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	return nil
}
