package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
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
		Update: resourceDcosPackageUpdate,
		Delete: resourceDcosPackageDelete,

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
			"wait": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Instructs the resource provider to wait until the resource is ready before continuing",
			},
			"config": {
				Type:        schema.TypeString,
				Required:    true,
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

	return normalizePackageSpec(packageSpec)
}

/**
 * Normalizes the package specifications, ensuring that the configuration sections
 * are merged in the correct order.
 */
func normalizePackageSpec(packageSpec *packageConfigSpec) (*packageVersionSpec, map[string]map[string]interface{}, error) {
	// Extract default config from the schema. This is going to be the
	// base where we are merge the config later.
	//
	// NOTE: We are merging here and not on the individual configuration data
	//       resources because we must always apply the defaults *first*.
	//
	config, err := util.DefaultJSONFromSchema(packageSpec.Version.Schema)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to extract package defaults from it's configuration schema: %s", err.Error())
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

/**
 * Get the app description
 */
func getAppDescription(client *dcos.APIClient, appId string) (*dcos.CosmosServiceDescribeV1Response, error) {
	ctx := context.TODO()

	describeOpts := &dcos.ServiceDescribeOpts{
		CosmosServiceDescribeV1Request: optional.NewInterface(dcos.CosmosServiceDescribeV1Request{
			AppId: appId,
		}),
	}

	log.Printf("[DEBUG] Querying cosmos to describe service with appId='%s'", appId)
	pkg, resp, err := client.Cosmos.ServiceDescribe(ctx, describeOpts)
	log.Printf("[TRACE] HTTP Response: %v", util.GetVerboseCosmosError(err, resp))

	if err != nil {
		log.Printf("[WARN] Got service description error: %s", util.GetVerboseCosmosError(err, resp))
		return nil, fmt.Errorf(util.GetVerboseCosmosError(err, resp))
	}

	log.Printf("[TRACE] Got service description: %v, %d", pkg.ResolvedOptions, len(pkg.ResolvedOptions))
	if len(pkg.ResolvedOptions) > 0 {
		return &pkg, nil
	}

	// If the service is not yet ready, return nil as a package response
	return nil, nil
}

/**
 * Gets the service description with the specified app ID, and waits up to <timeountMin> minutes until it's ready
 */
func waitAndGetAppDescription(client *dcos.APIClient, appId string, timeoutMin int) (*dcos.CosmosServiceDescribeV1Response, error) {
	var describeResult *dcos.CosmosServiceDescribeV1Response

	err := resource.Retry(time.Duration(timeoutMin)*time.Minute, func() *resource.RetryError {
		pkg, err := getAppDescription(client, appId)
		if err != nil {
			log.Printf("[WARN] Breaking out of retry loop because of unrecoverable error")
			return resource.NonRetryableError(err)
		}

		if pkg == nil {
			log.Printf("[TRACE] Got `nil` as package description, still waiting for the app")
			return resource.RetryableError(fmt.Errorf("Service %s still initializing", appId))
		}

		describeResult = pkg
		return resource.NonRetryableError(nil)
	})
	if err != nil {
		return nil, err
	}

	return describeResult, nil
}

/**
 * Get the specified package description
 */
func getPackageDescription(client *dcos.APIClient, packageName string, packageVersion string) (*dcos.CosmosPackage, error) {
	ctx := context.TODO()

	// Get the installed versions
	localVarOptionals := &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName:    packageName,
			PackageVersion: packageVersion,
		}),
	}

	log.Printf("[DEBUG] Querying cosmos to describe package='%s', version='%s'", packageName, packageVersion)
	descResp, httpResp, err := client.Cosmos.PackageDescribe(ctx, localVarOptionals)
	log.Printf("[TRACE] HTTP Response: %v", httpResp)

	if httpResp.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Package was not found")
		return nil, fmt.Errorf("Package %s does not exist", packageName)
	}
	if err != nil {
		log.Printf("[WARN] Got general error: %s", err.Error())
		return nil, fmt.Errorf("Unable to query cosmos: %s", util.GetVerboseCosmosError(err, httpResp))
	}

	log.Printf("[TRACE] Got package description: %v", descResp)
	return &descResp.Package, nil
}

func resourceDcosPackageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	packageVersion, packageConfig, err := collectPackageConfiguration(d.Get("config").(string))
	if err != nil {
		return fmt.Errorf("Unable to parse package config: %s", err.Error())
	}

	// First, make sure the package exists on the cosmos registry
	_, err = getPackageDescription(client, packageVersion.Name, packageVersion.Version)
	if err != nil {
		return err
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

	log.Printf("[DEBUG] Installing package %s:%s using cosmos", packageVersion.Name, packageVersion.Version)
	log.Printf("[DEBUG] Using options: %s", util.PrintJSON(cosmosPackageInstallV1Request.Options))
	installedPkg, httpResp, err := client.Cosmos.PackageInstall(ctx, cosmosPackageInstallV1Request)
	log.Printf("[TRACE] HTTP Response: %v", httpResp)

	if err != nil {
		log.Printf("[WARN] Cosmos install error: %s", err.Error())
		return fmt.Errorf("Unable to install package %s:%s: %s",
			packageVersion.Name,
			packageVersion.Version,
			util.GetVerboseCosmosError(err, httpResp),
		)
	}
	log.Printf("[DEBUG] Installed Package: %v", installedPkg)

	// Make sure the app_id is always populated, since it's an optional field
	d.Set("app_id", installedPkg.AppId)

	// If we should wait for the service, do it now
	if d.Get("wait").(bool) {
		_, err := waitAndGetAppDescription(client, installedPkg.AppId, 5)
		if err != nil {
			return fmt.Errorf("Error while waiting for the app to become available: %s", err.Error())
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", packageVersion.Name, installedPkg.AppId))
	return resourceDcosPackageRead(d, meta)
}

func resourceDcosPackageRead(d *schema.ResourceData, meta interface{}) error {
	var err error
	var desc *dcos.CosmosServiceDescribeV1Response
	client := meta.(*dcos.APIClient)

	// If the app_id is missing, this resource is never created. Guard against
	// this case as early as possible.
	vString, appIDok := d.GetOk("app_id")
	if !appIDok {
		log.Printf("[WARN] Missing 'app_id'. Assuming the service is not installed")
		d.SetId("")
		return nil
	}
	appID := vString.(string)

	// We are going to wait for 5 minutes for the app to appear, just in case we
	// were very quick on the previous deployment
	if d.Get("wait").(bool) {
		desc, err = waitAndGetAppDescription(client, appID, 5)
	} else {
		desc, err = getAppDescription(client, appID)
	}
	if err != nil {
		return fmt.Errorf("Error while querying app status: %s", err.Error())
	}
	if desc == nil {
		return fmt.Errorf("App '%s' was not available. Consider using `wait=true`")
	}

	// Try our best to the `packageConfigSpec` and `packageVersionSpec`
	// residing solely on the information we have collected.
	versionSpec := &packageVersionSpec{
		Name:    desc.Package.Name,
		Version: desc.Package.Version,
		Schema:  desc.Package.Config,
	}
	packageSpec := &packageConfigSpec{
		Version: versionSpec,
		Config:  desc.UserProvidedOptions,
	}
	spec, err := serializePackageConfigSpec(packageSpec)
	if err != nil {
		return fmt.Errorf("Unable to serialize the package spec: %s", err.Error())
	}
	d.Set("config", spec)

	d.SetId(fmt.Sprintf("%s:%s", desc.Package.Name, appID))
	return nil
}

func resourceDcosPackageUpdate(d *schema.ResourceData, meta interface{}) error {
	var desc *dcos.CosmosServiceDescribeV1Response
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	appID := d.Get("app_id").(string)

	// Enable partial state change, in order to properly manipulate the config
	d.Partial(true)
	if d.HasChange("config") {

		iOld, iNew := d.GetChange("config")
		oldVer, oldConfig, err := collectPackageConfiguration(iOld.(string))
		if err != nil {
			return fmt.Errorf("Unable to parse previous configuration: %s", err.Error())
		}
		newVer, newConfig, err := collectPackageConfiguration(iNew.(string))
		if err != nil {
			return fmt.Errorf("Unable to parse new configuration: %s", err.Error())
		}

		// We should never reach this case, but cover it just to be safe. (the ID
		// contains the package version and therefore a package name change will
		// *always* result into a new resource)
		if newVer.Name != oldVer.Name {
			return fmt.Errorf("Internal error: Reached unexpected `update` life-cycle event when package names have changed")
		}

		// Check if the configuration has changed (defaults included)
		oldHash, err := util.HashDict(util.NestedToFlatMap(oldConfig))
		if err != nil {
			return fmt.Errorf("Unable to hash old package configuration: %s", err.Error())
		}
		newHash, err := util.HashDict(util.NestedToFlatMap(newConfig))
		if err != nil {
			return fmt.Errorf("Unable to hash new package configuration: %s", err.Error())
		}

		// Check for version and/or config changes
		if newVer.Version != oldVer.Version {
			// First of all, make sure that we can jump to the given version,
			// stating from our current version

			// We are going to wait for 5 minutes for the app to appear, just in case we
			// were very quick on the previous deployment
			if d.Get("wait").(bool) {
				desc, err = waitAndGetAppDescription(client, appID, 5)
			} else {
				desc, err = getAppDescription(client, appID)
			}
			if err != nil {
				return fmt.Errorf("Error while querying app status: %s", err.Error())
			}
			if desc == nil {
				return fmt.Errorf("App '%s' was not available. Consider using `wait=true`")
			}

			// Guard against state discrepancies
			if desc.Package.Version != oldVer.Version {
				return fmt.Errorf(
					"Terraform state indicates version '%s', but package version '%s' was installed",
					oldVer.Version, desc.Package.Version,
				)
			}

			// Check if we can upgrade/downgrade to the target version
			var verEnum []string
			verFound := false
			log.Printf("[DEBUG] Checking if package upgrades to: %s", newVer.Version)
			for _, ver := range desc.UpgradesTo {
				log.Printf("[TRACE] Checking %s", ver)
				if ver == newVer.Version {
					verFound = true
					break
				}
				verEnum = append(verEnum, ver)
			}
			if !verFound {
				log.Printf("[DEBUG] Checking if package downgrades to: %s", newVer.Version)
				for _, ver := range desc.DowngradesTo {
					log.Printf("[TRACE] Checking %s", ver)
					if ver == newVer.Version {
						verFound = true
						break
					}
					verEnum = append(verEnum, ver)
				}
			}

			// If nothing found, we cannot continue
			if !verFound {
				return fmt.Errorf(
					"Service '%s' cannot be upgraded to version '%s'. Possible options are: %s",
					appID, newVer.Version, strings.Join(verEnum, ", "),
				)
			}

			// All checks are passed, we are now ready to ask cosmos for
			// a service upgrade to the new version. Any possible new option changes
			// will be included in the same request.
			cosmosServiceUpdateV1Request := dcos.CosmosServiceUpdateV1Request{
				AppId:          appID,
				PackageName:    newVer.Name,
				PackageVersion: newVer.Version,
				Options:        util.NestedToFlatMap(newConfig),
			}

			log.Printf("[DEBUG] Updating package %s:%s to version %s using cosmos", oldVer.Name, oldVer.Version, newVer.Version)
			log.Printf("[DEBUG] Using options: %s", util.PrintJSON(cosmosServiceUpdateV1Request.Options))
			_, httpResp, err := client.Cosmos.ServiceUpdate(ctx, cosmosServiceUpdateV1Request)
			log.Printf("[TRACE] HTTP Response: %v", httpResp)
			if err != nil {
				return fmt.Errorf("Unable to update package %s: %s", appID, util.GetVerboseCosmosError(err, httpResp))
			}

		} else if oldHash != newHash {

			// Plac
			cosmosServiceUpdateV1Request := dcos.CosmosServiceUpdateV1Request{
				AppId:          appID,
				PackageName:    oldVer.Name,
				PackageVersion: oldVer.Version,
				Options:        util.NestedToFlatMap(newConfig),
			}

			log.Printf("[DEBUG] Updating package %s:%s configuration using cosmos", oldVer.Name, oldVer.Version)
			log.Printf("[DEBUG] Using options: %s", util.PrintJSON(cosmosServiceUpdateV1Request.Options))
			_, httpResp, err := client.Cosmos.ServiceUpdate(ctx, cosmosServiceUpdateV1Request)
			log.Printf("[TRACE] HTTP Response: %v", httpResp)
			if err != nil {
				return fmt.Errorf("Unable to update service %s: %s", appID, util.GetVerboseCosmosError(err, httpResp))
			}

		} else {
			log.Printf("[INFO] No changes to app version or configuration were detected")
		}

		d.SetPartial("config")
	}
	d.Partial(false)

	return resourceDcosPackageRead(d, meta)
}

func resourceDcosPackageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()
	appID := d.Get("app_id").(string)

	packageVersion, _, err := collectPackageConfiguration(d.Get("config").(string))
	if err != nil {
		return fmt.Errorf("Unable to parse package config: %s", err.Error())
	}

	cosmosPackageUninstallV1Request := dcos.CosmosPackageUninstallV1Request{
		AppId:       appID,
		PackageName: packageVersion.Name,
	}

	// Unisntall package package
	_, resp, err := client.Cosmos.PackageUninstall(ctx, cosmosPackageUninstallV1Request, nil)
	log.Printf("[TRACE] Cosmos.PackageUninstall - %v", resp)
	if err != nil {
		return fmt.Errorf("Unable to uninstall package: %s", util.GetVerboseCosmosError(err, resp))
	}

	// If instructed, wait until it no loger appears on the enumeration
	if d.Get("wait").(bool) {
		listOpts := &dcos.PackageListOpts{
			CosmosPackageListV1Request: optional.NewInterface(dcos.CosmosPackageListV1Request{
				AppId:       appID,
				PackageName: packageVersion.Name,
			}),
		}
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			lst, resp, err := client.Cosmos.PackageList(ctx, listOpts)
			log.Printf("[TRACE] Cosmos.PackageList - %v, lst: %#v", resp, lst)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf(util.GetVerboseCosmosError(err, resp)))
			}

			log.Printf("[TRACE] Cosmos.ServiceDescribe - Packages %v, %d", lst.Packages, len(lst.Packages))
			if len(lst.Packages) == 0 {
				d.SetId("")
				return resource.NonRetryableError(nil)
			}

			return resource.RetryableError(fmt.Errorf("AppID %s still uninstalling", appID))
		})
		if err != nil {
			return fmt.Errorf("Error while waiting for the service to be uninstalled: %s", err.Error())
		}
	}

	d.SetId("")
	return nil
}
