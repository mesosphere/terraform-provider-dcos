package dcos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/imdario/mergo"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

func resourceDcosPackage() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosPackageCreate,
		Read:   resourceDcosPackageRead,
		// Update: resourceDcosPackageUpdate,
		Delete: resourceDcosPackageDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if filepath.Clean("/"+old) == filepath.Clean("/"+new) {
						return true
					}

					return false
				},
				Description: "ID of the account is used by default",
			},
			"config": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The package configuration to use",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					log.Printf("[Trace] Comparing old %s  with new %s", old, new)

					oldVer, oldConfig, err := collectPackageConfiguration(old)
					if err != nil {
						log.Printf("[WARNING] Unable to parse old package config: %s", err.Error())
						return false
					}

					newVer, newConfig, err := collectPackageConfiguration(new)
					if err != nil {
						log.Printf("[WARNING] Unable to parse new package config: %s", err.Error())
						return false
					}

					// Check if the version specifications have changed
					if oldVer.Name != newVer.Name {
						log.Printf("[TRACE] Old name '%s' does not match new name '%s'", oldVer.Name, newVer.Name)
						return false
					}
					if oldVer.Version != newVer.Version {
						log.Printf("[TRACE] Old version '%s' does not match new version '%s'", oldVer.Version, newVer.Version)
						return false
					}

					// Check if the configuration has changed (defaults included)
					oldHash, err := util.HashDict(util.NestedToFlatMap(oldConfig))
					if err != nil {
						log.Printf("[WARNING] Unable to hash old package config: %s", err.Error())
						return false
					}
					newHash, err := util.HashDict(util.NestedToFlatMap(newConfig))
					if err != nil {
						log.Printf("[WARNING] Unable to hash new package config: %s", err.Error())
						return false
					}
					if oldHash != newHash {
						log.Printf("[TRACE] Old config hash '%s' does not match new config hash '%s'", oldHash, newHash)
						return false
					}

					return true
				},
			},
		},
	}
}

/**
 * Collects the package configuration. This includes:
 *
 * - Getting the defaults from the version spec
 * - Merging it with the configuration JSON
 */
func collectPackageConfiguration(configSpec string) (*packageVersionSpec, map[string]map[string]interface{}, error) {
	packageSpec, err := deserializePackageConfigSpec(configSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to parse configuration: %s", err.Error())
	}
	if packageSpec.Version == nil {
		return nil, nil, fmt.Errorf("The configuration given do not include a version spec")
	}

	// Extract default config from the schema. This is going to be the
	// base where we are merge the config later.
	//
	// NOTE: We are merging here and not on the individual configuration data
	//       resources because we must always apply the defaults *first*.
	//
	config, err := util.DefaultJSONFromSchema(packageSpec.Version.Schema)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to extract package defaults from it's configuration schema")
	}

	// Merge package configuration
	pkgConfig, err := util.FlatToNestedMap(packageSpec.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("Unexpected error in the configuration: %s", err.Error())
	}
	err = mergo.MergeWithOverwrite(&config, &pkgConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to merge the configuration with the defaults: %s", err.Error())
	}

	// Done
	return packageSpec.Version, config, nil
}

func resourceDcosPackageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	packageVersion, packageConfig, err := collectPackageConfiguration(d.Get("config").(string))
	if err != nil {
		return fmt.Errorf("Unable to parse package config: %s", err.Error())
	}

	// First, make sure the package exists on the cosmos registry
	localVarOptionals := &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName: packageVersion.Name,
		}),
	}
	_, resp, err := client.Cosmos.PackageDescribe(ctx, localVarOptionals)
	log.Printf("[TRACE] Cosmos.PackageDescribe - %v", resp)
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Package %s does not exist", packageVersion.Name)
	}
	if err != nil {
		return fmt.Errorf("Unable to query cosmos: %s", err.Error())
	}

	// Prepare for package install
	cosmosPackageInstallV1Request := dcos.CosmosPackageInstallV1Request{}
	cosmosPackageInstallV1Request.PackageName = packageVersion.Name
	if appID, ok := d.GetOk("app_id"); ok {
		cosmosPackageInstallV1Request.AppId = appID.(string)
	} else {
		cosmosPackageInstallV1Request.AppId = packageVersion.Name
		d.Set("app_id", appID.(string))
	}
	cosmosPackageInstallV1Request.PackageVersion = packageVersion.Version
	cosmosPackageInstallV1Request.Options = util.NestedToFlatMap(packageConfig)

	installedPkg, installResp, err := client.Cosmos.PackageInstall(ctx, cosmosPackageInstallV1Request)
	log.Printf("[TRACE] Cosmos.PackageInstall - %v", installResp)

	if err != nil {
		return err
	}

	log.Printf("[INFO] Installed Package - %v", installedPkg)

	d.Set("app_id", installedPkg.AppId)
	d.SetId(installedPkg.AppId)

	return resourceDcosPackageRead(d, meta)
}

func resourceDcosPackageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	var appID string
	a, appIDok := d.GetOk("app_id")
	if !appIDok {
		d.SetId("")
		return nil
	}
	appID = a.(string)

	describeOpts := &dcos.ServiceDescribeOpts{
		CosmosServiceDescribeV1Request: optional.NewInterface(dcos.CosmosServiceDescribeV1Request{
			AppId: appID,
		}),
	}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		pkg, resp, err := client.Cosmos.ServiceDescribe(ctx, describeOpts)

		log.Printf("[TRACE] Cosmos.ServiceDescribe - %v, pkg: %#v", resp, pkg)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		log.Printf("[TRACE] Cosmos.ServiceDescribe - ResolvedOptions %v, %d", pkg.ResolvedOptions, len(pkg.ResolvedOptions))

		if len(pkg.ResolvedOptions) > 0 {
			d.Set("version", pkg.Package.Version)
			if configJSON, err := json.Marshal(pkg.ResolvedOptions); err == nil {
				d.Set("config_json", string(configJSON))
			}
			d.Set("name", pkg.Package.Name)
			d.SetId(appID)

			return resource.NonRetryableError(nil)
		} else {
			return resource.RetryableError(fmt.Errorf("AppID %s still initializing", appID))
		}
	})
}

func resourceDcosPackageUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*dcos.APIClient)
	// ctx := context.TODO()

	return resourceDcosPackageRead(d, meta)
}

func resourceDcosPackageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	appID := d.Get("app_id").(string)
	packageName := d.Get("name").(string)

	cosmosPackageUninstallV1Request := dcos.CosmosPackageUninstallV1Request{
		AppId:       appID,
		PackageName: packageName,
	}

	// Uninstall package
	_, resp, err := client.Cosmos.PackageUninstall(ctx, cosmosPackageUninstallV1Request, nil)
	log.Printf("[TRACE] Cosmos.PackageUninstall - %v", resp)
	if err != nil {
		return fmt.Errorf("Unable to uninstall package: %s", err.Error())
	}

	// Wait until it does no longer appear on the enumeration
	listOpts := &dcos.PackageListOpts{
		CosmosPackageListV1Request: optional.NewInterface(dcos.CosmosPackageListV1Request{
			AppId:       appID,
			PackageName: packageName,
		}),
	}
	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		lst, resp, err := client.Cosmos.PackageList(ctx, listOpts)
		log.Printf("[TRACE] Cosmos.PackageList - %v, lst: %#v", resp, lst)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		log.Printf("[TRACE] Cosmos.ServiceDescribe - Packages %v, %d", lst.Packages, len(lst.Packages))
		if len(lst.Packages) == 0 {
			d.SetId("")
			return resource.NonRetryableError(nil)
		}

		return resource.RetryableError(fmt.Errorf("AppID %s still uninstalling", appID))
	})

	d.SetId("")
	return nil
}
