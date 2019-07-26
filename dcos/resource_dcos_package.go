package dcos

import (
	"context"
	"encoding/json"
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
			"wait_duration": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "5m",
				Description: "The duration to wait for a deployment or teardown to complete",
			},
			"sdk": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enables SDK-specific APIs for this package",
			},
			"config": schemaInPackageConfigSpecWithDiffSup(),
		},
	}
}

/**
 * stripRootSlash removes the leading '/' from App names
 */
func stripRootSlash(appId string) string {
	// Trim leading slash from appId
	serviceName := appId
	if strings.HasPrefix(serviceName, "/") {
		serviceName = appId[1:]
	}

	return serviceName
}

/**
 * schemaInPackageConfigSpecWithDiffSup extends the schema returned by `schemaInPackageConfigSpec`,
 * by adding a diff suppression function that checks if the user-given options are different
 * than the options returned by the service during the Read lifecycle.
 */
func schemaInPackageConfigSpecWithDiffSup() *schema.Schema {
	baseSchema := schemaInPackageConfigSpec(true)
	baseSchema.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
		log.Printf("[TRACE] Comparing old %s '%s' <== with new ==> '%s'", k, old, new)
		switch k {
		case "config.%":
			return false

		case "config.config":

			if new == "" {
				log.Printf("[DEBUG] New config is blank, assuming no changed")
				return true
			}
			if old == "" {
				log.Printf("[DEBUG] Old config is blank, assuming changed")
				return false
			}

			// If we cannot parse the contents, assume there are differences, so don't
			// suppress the configuration.
			savedMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(old), &savedMap)
			if err != nil {
				log.Printf("[WARN] Unable to parse old package config: %s", old)
				return false
			}
			userMap := make(map[string]interface{})
			err = json.Unmarshal([]byte(new), &userMap)
			if err != nil {
				log.Printf("[WARN] Unable to parse new package config: %s", new)
				return false
			}

			// Check if whatever options are given from the new configuration are actually
			// changing something on the saved map.
			diff := util.GetDictDiff(savedMap, userMap)
			log.Printf("[DEBUG] Delta between saved and user map: %s", util.PrintJSON(diff))
			if len(diff) == 0 {
				return true
			}
			return false

		default:
			eq := old == new
			log.Printf("[DEBUG] Equality: %v", eq)
			return eq
		}

		return false
	}

	return baseSchema
}

/**
 * updateServiceName updates in-place the options map with the correct service name
 */
func updateServiceName(options map[string]interface{}, appId string) {
	var serviceMap map[string]interface{}

	// Create "service" section
	if v, ok := options["service"]; ok {
		if vMap, ok := v.(map[string]interface{}); ok {
			serviceMap = vMap
		} else {
			serviceMap = make(map[string]interface{})
			options["service"] = serviceMap
		}
	} else {
		serviceMap = make(map[string]interface{})
		options["service"] = serviceMap
	}

	// Update service name
	serviceMap["name"] = stripRootSlash(appId)
}

/**
 * collectPackageConfiguration breaks down the given configSpec into the package version
 * and the normalized configuration.
 */
func collectPackageConfiguration(configSpec map[string]interface{}) (*packageVersionSpec, string, map[string]map[string]interface{}, error) {
	packageSpec, err := deserializePackageConfigSpec(configSpec)
	if err != nil {
		return nil, "", nil, fmt.Errorf("Unable to parse configuration: %s", err.Error())
	}
	if packageSpec.Version == nil {
		return nil, "", nil, fmt.Errorf("The configuration given do not include a version spec")
	}

	return normalizePackageSpec(packageSpec)
}

/**
 * normalizePackageSpec "normalizes" the package specifications, ensuring that
 * the configuration sections are merged in the correct order and defaults are
 * included.
 */
func normalizePackageSpec(packageSpec *packageConfigSpec) (*packageVersionSpec, string, map[string]map[string]interface{}, error) {
	// Extract default config from the schema. This is going to be the
	// base where we are merge the config later.
	//
	// NOTE: We are merging the data here in the resource and *not* on the
	//       individual package_config data resources because we must always
	// 			 apply the defaults first.
	//
	config, err := util.DefaultJSONFromSchema(packageSpec.Version.Schema)
	if err != nil {
		return nil, "", nil, fmt.Errorf("Unable to extract package defaults from it's configuration schema: %s", err.Error())
	}

	// Merge package configuration
	pkgConfig, err := util.FlatToNestedMap(packageSpec.Config)
	if err != nil {
		return nil, "", nil, fmt.Errorf("Unexpected error in the configuration: %s", err.Error())
	}
	err = mergo.MergeWithOverwrite(&config, &pkgConfig)
	if err != nil {
		return nil, "", nil, fmt.Errorf("Unable to merge the configuration with the defaults: %s", err.Error())
	}

	// Done
	return packageSpec.Version, packageSpec.Checksum, config, nil
}

/**
 * getPackageSpecFromServiceDesc converts the given cosmos service description
 * response into a `packageConfigSpec`
 */
func getPackageSpecFromServiceDesc(desc *dcos.CosmosServiceDescribeV1Response) *packageConfigSpec {
	// Try our best to the `packageConfigSpec` and `packageVersionSpec`
	// residing solely on the information we have collected.
	versionSpec := &packageVersionSpec{
		Name:    desc.Package.Name,
		Version: desc.Package.Version,
		Schema:  desc.Package.Config,
	}
	return &packageConfigSpec{
		Version: versionSpec,
		Config:  desc.UserProvidedOptions,
	}
}

/**
 * getServiceDesc queries cosmos and returns a service description for the given
 * app ID. This response includes the package-specific details typically obtained
 * via getPackageDesc.
 */
func getServiceDesc(client *dcos.APIClient, appId string) (*dcos.CosmosServiceDescribeV1Response, error) {
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
		if apiError, ok := err.(dcos.GenericOpenAPIError); ok {
			if apiError.Model() != nil {
				if cosmosError, ok := apiError.Model().(dcos.CosmosError); ok {
					if cosmosError.Type == "MarathonAppNotFound" {
						return nil, nil
					}
				}
			}
		}
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
 * waitAndgetServiceDesc gets the service description with the specified app ID,
 * and waits up to <timeountMin> minutes until it's ready.
 */
func waitAndgetServiceDesc(client *dcos.APIClient, appId string, timeout time.Duration) (*dcos.CosmosServiceDescribeV1Response, error) {
	var describeResult *dcos.CosmosServiceDescribeV1Response

	err := resource.Retry(timeout, func() *resource.RetryError {
		pkg, err := getServiceDesc(client, appId)
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
 * waitForSDKPlan keeps querying for the plan specified and waits until it enters
 * the given status, or the timeout event occurs.
 */
func waitForSDKPlan(client *dcos.APIClient, appId string, planName string, waitStatus string, timeout time.Duration) error {
	sdkClient := util.CreateSDKAPIClient(client, appId)

	return resource.Retry(timeout, func() *resource.RetryError {
		plan, err := sdkClient.PlanGetStatus(planName)
		if err != nil {
			log.Printf("[WARN] Error querying plan %s status: %s", planName, err.Error())
			return resource.RetryableError(
				fmt.Errorf("Service %s is not yet responding", appId),
			)
		}

		log.Printf("[TRACE] Got plan response: %v", plan)
		if plan.Status != waitStatus {
			return resource.RetryableError(fmt.Errorf(
				"Service plan %s is %s (expecting %s)",
				appId, planName, plan.Status, waitStatus,
			))
		}

		return resource.NonRetryableError(nil)
	})
}

/**
 * getPackageDesc queries cosmos for a package with the given name or version
 * and returns the package details.
 * `packageVersion` can be blank if you are querying for the latest version.
 */
func getPackageDesc(client *dcos.APIClient, packageName string, packageVersion string) (*dcos.CosmosPackage, error) {
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

	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
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

/**
 * resourceDcosPackageCreate is the default resource `Create` handler
 */
func resourceDcosPackageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	packageVersion, configCsum, packageConfig, err := collectPackageConfiguration(d.Get("config").(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("Unable to parse package config: %s", err.Error())
	}

	// First, make sure the package exists on the cosmos registry
	_, err = getPackageDesc(client, packageVersion.Name, packageVersion.Version)
	if err != nil {
		return err
	}

	// Then check if a similar application already exists
	appId := packageVersion.Name
	if v, ok := d.GetOk("app_id"); ok {
		appId = stripRootSlash(v.(string))
	}
	log.Printf("[TRACE] CREATE Lifecycle - app %s", appId)

	// TODO: Check if installed app and new app spec matches

	// Prepare for package install
	cosmosPackageInstallV1Request := dcos.CosmosPackageInstallV1Request{}
	cosmosPackageInstallV1Request.PackageName = packageVersion.Name
	cosmosPackageInstallV1Request.PackageVersion = packageVersion.Version
	cosmosPackageInstallV1Request.AppId = appId
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
	d.Set("app_id", stripRootSlash(installedPkg.AppId))

	// If we should wait for the service, do it now
	if d.Get("wait").(bool) {
		waitDuration, err := time.ParseDuration(d.Get("wait_duration").(string))
		if err != nil {
			return fmt.Errorf("Unable to parse wait duration")
		}

		_, err = waitAndgetServiceDesc(client, installedPkg.AppId, waitDuration)
		if err != nil {
			return fmt.Errorf("Error while waiting for the app to become available: %s", err.Error())
		}

		// If this is an SDK service, also wait for the deployment plan to be completed
		if d.Get("sdk").(bool) {
			err = waitForSDKPlan(client, appId, "deploy", "COMPLETE", waitDuration)
			if err != nil {
				return fmt.Errorf("Error while waiting for the deployment plan to complete: %s", err.Error())
			}
		}
	}

	// Keep track of a configuration ID
	sdkClient := util.CreateSDKAPIClient(client, appId)
	sdkClient.SetMeta("csum", configCsum)

	d.SetId(fmt.Sprintf("%s:%s", packageVersion.Name, installedPkg.AppId))
	return resourceDcosPackageRead(d, meta)
}

/**
 * resourceDcosPackageRead is the default resource `Read` handler
 */
func resourceDcosPackageRead(d *schema.ResourceData, meta interface{}) error {
	var err error
	var desc *dcos.CosmosServiceDescribeV1Response
	client := meta.(*dcos.APIClient)

	// If the app_id is missing, this resource is never created. Guard against
	// this case as early as possible.
	vString, appIdok := d.GetOk("app_id")
	if !appIdok {
		log.Printf("[WARN] Missing 'app_id'. Assuming the service is not installed")
		d.SetId("")
		return nil
	}
	appId := stripRootSlash(vString.(string))
	log.Printf("[TRACE] READ Lifecycle - app %s", appId)

	sdkClient := util.CreateSDKAPIClient(client, appId)

	// We are going to wait for 5 minutes for the app to appear, just in case we
	// were very quick on the previous deployment
	desc, err = getServiceDesc(client, appId)
	if err != nil {
		return fmt.Errorf("Error while querying app status: %s", err.Error())
	}
	if desc == nil {
		d.SetId("")
		return nil
	}

	// Query SDK meta to get the old config checksum
	csum, err := sdkClient.GetMeta("csum", "")
	if err != nil {
		return fmt.Errorf("Error fetching old config checksum: %s", err.Error())
	}

	// Compute package spec from the service description
	packageSpec := getPackageSpecFromServiceDesc(desc)
	packageSpec.Checksum = csum.(string)

	// Serialize and store the new config
	spec, err := serializePackageConfigSpec(packageSpec)
	if err != nil {
		return fmt.Errorf("Unable to serialize the package spec: %s", err.Error())
	}
	d.Set("config", spec)

	d.SetId(fmt.Sprintf("%s:%s", desc.Package.Name, appId))
	return nil
}

/**
 * resourceDcosPackageUpdate is the default resource `Update` handler
 */
func resourceDcosPackageUpdate(d *schema.ResourceData, meta interface{}) error {
	var desc *dcos.CosmosServiceDescribeV1Response
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	appId := stripRootSlash(d.Get("app_id").(string))
	sdkClient := util.CreateSDKAPIClient(client, appId)
	log.Printf("[TRACE] UPDATE Lifecycle - app %s", appId)

	// Enable partial state change, in order to properly manipulate the config
	d.Partial(true)
	if d.HasChange("config") {

		iOld, iNew := d.GetChange("config")
		oldVer, oldChecksum, oldConfig, err := collectPackageConfiguration(iOld.(map[string]interface{}))
		if err != nil {
			return fmt.Errorf("Unable to parse previous configuration: %s", err.Error())
		}
		newVer, newChecksum, newConfig, err := collectPackageConfiguration(iNew.(map[string]interface{}))
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
				waitDuration, err := time.ParseDuration(d.Get("wait_duration").(string))
				if err != nil {
					return fmt.Errorf("Unable to parse wait duration")
				}
				desc, err = waitAndgetServiceDesc(client, appId, waitDuration)
			} else {
				desc, err = getServiceDesc(client, appId)
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
					appId, newVer.Version, strings.Join(verEnum, ", "),
				)
			}

			// All checks are passed, we are now ready to ask cosmos for
			// a service upgrade to the new version. Any possible new option changes
			// will be included in the same request.
			cosmosServiceUpdateV1Request := dcos.CosmosServiceUpdateV1Request{
				AppId:          appId,
				PackageName:    newVer.Name,
				PackageVersion: newVer.Version,
				Options:        util.NestedToFlatMap(newConfig),
			}

			// Ensure that the service name points to the app ID
			updateServiceName(cosmosServiceUpdateV1Request.Options, appId)

			log.Printf("[DEBUG] Updating package %s:%s to version %s using cosmos", oldVer.Name, oldVer.Version, newVer.Version)
			log.Printf("[DEBUG] Using options: %s", util.PrintJSON(cosmosServiceUpdateV1Request.Options))
			_, httpResp, err := client.Cosmos.ServiceUpdate(ctx, cosmosServiceUpdateV1Request)
			log.Printf("[TRACE] HTTP Response: %v", httpResp)
			if err != nil {
				return fmt.Errorf("Unable to update package %s: %s", appId, util.GetVerboseCosmosError(err, httpResp))
			}

			// If this is an SDK service, also wait for the deployment plan to be completed
			if d.Get("wait").(bool) && d.Get("sdk").(bool) {
				waitDuration, err := time.ParseDuration(d.Get("wait_duration").(string))
				if err != nil {
					return fmt.Errorf("Unable to parse wait duration")
				}

				err = waitForSDKPlan(client, appId, "deploy", "COMPLETE", waitDuration)
				if err != nil {
					return fmt.Errorf("Error while waiting for the deployment plan to complete: %s", err.Error())
				}
			}

		} else if oldHash != newHash {
			log.Printf("[INFO] Configuration has changed. Going to re-deploy")

			// If the configuration has changed, do not supply a new package version
			// but do suupply the new configuration.
			cosmosServiceUpdateV1Request := dcos.CosmosServiceUpdateV1Request{
				AppId:       appId,
				PackageName: oldVer.Name,
				Options:     util.NestedToFlatMap(newConfig),
			}

			// Ensure that the service name points to the app ID
			updateServiceName(cosmosServiceUpdateV1Request.Options, appId)

			log.Printf("[DEBUG] Updating package %s:%s configuration using cosmos", oldVer.Name, oldVer.Version)
			log.Printf("[DEBUG] Using options: %s", util.PrintJSON(cosmosServiceUpdateV1Request.Options))
			_, httpResp, err := client.Cosmos.ServiceUpdate(ctx, cosmosServiceUpdateV1Request)
			log.Printf("[TRACE] HTTP Response: %v", httpResp)
			if err != nil {
				return fmt.Errorf("Unable to update service %s: %s", appId, util.GetVerboseCosmosError(err, httpResp))
			}

			// If this is an SDK service, also wait for the deployment plan to be completed
			if d.Get("wait").(bool) && d.Get("sdk").(bool) {
				waitDuration, err := time.ParseDuration(d.Get("wait_duration").(string))
				if err != nil {
					return fmt.Errorf("Unable to parse wait duration")
				}

				err = waitForSDKPlan(client, appId, "deploy", "COMPLETE", waitDuration)
				if err != nil {
					return fmt.Errorf("Error while waiting for the deployment plan to complete: %s", err.Error())
				}
			}

		} else if oldChecksum != newChecksum {
			log.Printf("[INFO] Configuration and version is identical, but checksum has changed. Going to restart")

			// We can only gracefully restart SDK services, by restarting the deploy plan.
			// Standard cosmos packages have to be un-installed and re-installed.
			if d.Get("sdk").(bool) {
				err := sdkClient.PlanRestart("deploy")
				if err != nil {
					return fmt.Errorf("Unable to restart 'deploy' plan: %s", err.Error())
				}
			} else {
				log.Printf("[WARN] We currently don't support restarting non-SDK services!")
				// TODO: Remove and re-install service
			}

		} else {
			log.Printf("[INFO] No changes to app version or configuration were detected")
		}

		// Update the configuration checksum
		sdkClient.SetMeta("csum", newChecksum)

		d.SetPartial("config")
	}
	d.Partial(false)

	return resourceDcosPackageRead(d, meta)
}

/**
 * resourceDcosPackageDelete is the default resource `Delete` handler
 */
func resourceDcosPackageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()
	appId := stripRootSlash(d.Get("app_id").(string))
	log.Printf("[TRACE] DELETE Lifecycle - app %s", appId)

	// We are going to get reaped by the SDK uninstall, but just in case
	sdkClient := util.CreateSDKAPIClient(client, appId)
	sdkClient.SetMeta("csum", "")

	packageVersion, _, _, err := collectPackageConfiguration(d.Get("config").(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("Unable to parse package config: %s", err.Error())
	}

	cosmosPackageUninstallV1Request := dcos.CosmosPackageUninstallV1Request{
		AppId:       appId,
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
		log.Printf("[TRACE] Waiting for removal of app %s", appId)
		listOpts := &dcos.PackageListOpts{
			CosmosPackageListV1Request: optional.NewInterface(dcos.CosmosPackageListV1Request{
				AppId:       appId,
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

			return resource.RetryableError(fmt.Errorf("appId %s still uninstalling", appId))
		})
		if err != nil {
			return fmt.Errorf("Error while waiting for the service to be uninstalled: %s", err.Error())
		}
	}

	d.SetId("")
	return nil
}
