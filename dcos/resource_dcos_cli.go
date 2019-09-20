package dcos

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/imdario/mergo"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

var rxIDArg = regexp.MustCompile(`%NAME%`)

func resourceDcosCLI() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosCLICreate,
		Read:   resourceDcosCLIRead,
		Update: resourceDcosCLIUpdate,
		Delete: resourceDcosCLIDelete,

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
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
				Computed:    true,
				Optional:    true,
				Description: "The configuration JSON to carry along",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					id := d.Get("name").(string)
					old = rxIDArg.ReplaceAllString(old, id)
					new = rxIDArg.ReplaceAllString(new, id)

					a, err := util.NormalizeJSON(old)
					if err != nil {
						return false
					}

					b, err := util.NormalizeJSON(new)
					if err != nil {
						return false
					}

					return a == b
				},
			},

			"args_create": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The command to invoke for creating the resource",
			},
			"args_delete": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The command to invoke for deleting the resource",
			},
			"args_read": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The command to invoke for reading the resource",
			},
			"args_update": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The command to invoke for updating the resource",
			},

			"probe_wait_duration": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "5m",
				Description: "How long to wait for a probe to complete",
			},
			"probe_created": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The probe script to run to check if the service is created",
			},
			"probe_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The probe script to run to check if the service is updated",
			},
			"probe_deleted": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The probe script to run to check if the service is deleted",
			},
		},
	}
}

func waitProbe(d *schema.ResourceData, meta interface{}, probeName string) error {
	id := d.Get("name").(string)
	w := meta.(*ProviderState).CliWrapper
	pkgCli, err := w.ForPackage(d.Get("package").(string), id)
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	waitDuration, err := time.ParseDuration(d.Get("probe_wait_duration").(string))
	if err != nil {
		return fmt.Errorf("Unable to parse 'probe_wait_duration' value")
	}

	probeScript := d.Get(probeName).(string)
	probeScript = rxIDArg.ReplaceAllString(probeScript, id)
	if probeScript == "" {
		return nil
	}

	return resource.Retry(waitDuration, func() *resource.RetryError {
		err = pkgCli.Exec(probeScript, true)
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return resource.RetryableError(fmt.Errorf("Probe exited with non-zero code"))
			}
			return resource.NonRetryableError(err)
		}

		// If probe passes, things are good
		return resource.NonRetryableError(nil)
	})
}

func parseUserConfig(d *schema.ResourceData) (map[string]interface{}, error) {
	var configMap map[string]interface{}

	id := d.Get("name").(string)
	config := d.Get("config").(string)
	config = rxIDArg.ReplaceAllString(config, id)

	err := json.Unmarshal([]byte(config), &configMap)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse configuration spec '%s': %s", config, err.Error())
	}

	return configMap, nil
}

func resourceDcosCLICreate(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)
	w := meta.(*ProviderState).CliWrapper
	pkgCli, err := w.ForPackage(d.Get("package").(string), id)
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	// Read user configuration
	config, err := parseUserConfig(d)
	if err != nil {
		return fmt.Errorf("Unable to parse user config: %s", err.Error())
	}

	pkgCli.Config = config
	log.Printf("Creating id='%s', config=%s", id, util.PrintJSON(config))
	err = pkgCli.Exec(d.Get("args_create").(string), false)
	if err != nil {
		if _, ok := err.(*util.CliWrapperConfigParseError); ok {
			// Ignore
		} else {
			return fmt.Errorf("Unable to create resource: %s: %s", err.Error(), pkgCli.LastOutput)
		}
	}

	// Wait for resource
	err = waitProbe(d, meta, "probe_created")
	if err != nil {
		return fmt.Errorf("Error while waiting for resource to become available: %s", err.Error())
	}

	// Read resource
	err = resourceDcosCLIRead(d, meta)
	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceDcosCLIRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)
	w := meta.(*ProviderState).CliWrapper
	pkgCli, err := w.ForPackage(d.Get("package").(string), id)
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	// Read user configuration
	config, err := parseUserConfig(d)
	if err != nil {
		return fmt.Errorf("Unable to parse user config: %s", err.Error())
	}

	// Read remote config
	err = pkgCli.Exec(d.Get("args_read").(string), false)
	if err != nil {
		// If we were not able to parse the config, it means we have lost
		// the resource.
		log.Printf("Could not read, assuming deleted")
		d.SetId("")
		return nil
	}

	log.Printf("Read config=%s", id, util.PrintJSON(pkgCli.Config))

	// Calculate differences
	diff := util.GetDictDiff(pkgCli.Config, config)
	if len(diff) != 0 {
		err = mergo.MergeWithOverwrite(&config, &diff)
		if err != nil {
			return fmt.Errorf("Unable to apply config diff: %s", err.Error())
		}
	}

	log.Printf("Found diff=%s", id, util.PrintJSON(diff))

	sConfig, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("Unable to marshal the new config: %s", err.Error())
	}

	log.Printf("Computed new config=%s", id, util.PrintJSON(sConfig))

	// Apply new config
	d.SetId(id)
	d.Set("config", string(sConfig))

	return nil
}

func resourceDcosCLIUpdate(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)
	w := meta.(*ProviderState).CliWrapper
	pkgCli, err := w.ForPackage(d.Get("package").(string), id)
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	// Read user configuration
	config, err := parseUserConfig(d)
	if err != nil {
		return fmt.Errorf("Unable to parse user config: %s", err.Error())
	}

	pkgCli.Config = config
	log.Printf("Updating id='%s', config=%s", id, util.PrintJSON(config))
	err = pkgCli.Exec(d.Get("args_update").(string), false)
	if err != nil {
		if _, ok := err.(*util.CliWrapperConfigParseError); ok {
			// Ignore
		} else {
			return fmt.Errorf("Unable to update resource: %s: %s", err.Error(), pkgCli.LastOutput)
		}
	}

	// Wait for resource
	err = waitProbe(d, meta, "probe_updated")
	if err != nil {
		return fmt.Errorf("Error while waiting for resource to be updated: %s", err.Error())
	}

	// Read resource
	return resourceDcosCLIRead(d, meta)
}

func resourceDcosCLIDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Get("name").(string)
	w := meta.(*ProviderState).CliWrapper
	pkgCli, err := w.ForPackage(d.Get("package").(string), id)
	if err != nil {
		return fmt.Errorf("Unable to obtain package cli: %s", err.Error())
	}

	// Read user configuration
	config, err := parseUserConfig(d)
	if err != nil {
		return fmt.Errorf("Unable to parse user config: %s", err.Error())
	}

	log.Printf("Deleting id='%s'", id)

	pkgCli.Config = config
	log.Printf("Deleting id='%s', config=%s", id, util.PrintJSON(config))
	err = pkgCli.Exec(d.Get("args_delete").(string), false)
	if err != nil {
		if _, ok := err.(*util.CliWrapperConfigParseError); ok {
			// Ignore
		} else {
			return fmt.Errorf("Unable to delete resource: %s: %s", err.Error(), pkgCli.LastOutput)
		}
	}

	// Wait for resource
	err = waitProbe(d, meta, "probe_deleted")
	if err != nil {
		return fmt.Errorf("Error while waiting for resource to become unavailable: %s", err.Error())
	}

	// Read resource
	d.SetId("")
	return nil
}
