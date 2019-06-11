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

	"github.com/go-test/deep"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/imdario/mergo"
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Description of the newly created service account",
			},
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
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Description of the newly created service account",
			},

			"config_json": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					log.Printf("[Trace] Comparing old %s  with new %s", old, new)
					jconfig := make(map[string]map[string]interface{})

					oldConfig, err := unmarshallPackageConfig(old)
					if err != nil {
						return false
					}

					config, err := unmarshallPackageConfig(old)
					if err != nil {
						return false
					}

					jconfig, err = unmarshallPackageConfig(new)
					if err != nil {
						return false
					}

					err = mergo.MergeWithOverwrite(&config, &jconfig)
					if err != nil {
						log.Printf("[WARNING] config_json.DiffSuppressFunc Merge - %v", err)
						return false
					}
					log.Printf("[Trace] Comparing old %#v  with new %#v", oldConfig, config)

					if diff := deep.Equal(&oldConfig, &config); diff != nil {
						log.Printf("[TRACE] config_json.DiffSuppressFunc Equal - %s", strings.Join(diff, ","))

						return false
					}

					return true
				},
				Description: "Path to public key to use",
			},
		},
	}
}

func unmarshallPackageConfig(j string) (map[string]map[string]interface{}, error) {
	config := make(map[string]map[string]interface{})
	if err := json.Unmarshal([]byte(j), &config); err != nil {
		log.Printf("[WARNING] config_json.DiffSuppressFunc Unmarschal - %v", err)
		return nil, err
	}

	return config, nil
}

func resourceDcosPackageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	packageName := d.Get("name").(string)

	localVarOptionals := &dcos.PackageDescribeOpts{
		CosmosPackageDescribeV1Request: optional.NewInterface(dcos.CosmosPackageDescribeV1Request{
			PackageName: packageName,
		}),
	}
	_, resp, err := client.Cosmos.PackageDescribe(ctx, localVarOptionals)
	log.Printf("[TRACE] Cosmos.PackageDescribe - %v", resp)
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("package %s does not exist", packageName)
	}

	if err != nil {
		return err
	}

	// prepare packe install

	cosmosPackageInstallV1Request := dcos.CosmosPackageInstallV1Request{}
	cosmosPackageInstallV1Request.PackageName = packageName
	if appID, ok := d.GetOk("app_id"); ok {
		cosmosPackageInstallV1Request.AppId = appID.(string)
	} else {
		cosmosPackageInstallV1Request.AppId = packageName
		d.Set("app_id", appID.(string))
	}
	if packageVersion, ok := d.GetOk("version"); ok {
		cosmosPackageInstallV1Request.PackageVersion = packageVersion.(string)
	}

	if packageConfig, ok := d.GetOk("config_json"); ok {
		var opt map[string]map[string]interface{}
		err := json.Unmarshal([]byte(packageConfig.(string)), &opt)
		if err != nil {
			return fmt.Errorf("Error reading config_json %v", err)
		}
		cosmosPackageInstallV1Request.Options = opt

		log.Printf("[TRACE] Prepare Cosmos.PackageInstall found config_json - %v", opt)
	}

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
