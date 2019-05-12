package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "ID of the account is used by default",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Description of the newly created service account",
			},

			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "Path to public key to use",
			},
		},
	}
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

	pkgName := d.Get("name").(string)

	listOpts := &dcos.PackageListOpts{
		CosmosPackageListV1Request: optional.NewInterface(dcos.CosmosPackageListV1Request{
			AppId:       appID,
			PackageName: pkgName,
		}),
	}

	packages, resp, err := client.Cosmos.PackageList(ctx, listOpts)

	log.Printf("[TRACE] Cosmos.PackageList - %v", resp)

	if err != nil {
		return err
	}

	log.Printf("[TRACE] Cosmos.PackageList found %d packages", len(packages.Packages))

	for _, p := range packages.Packages {
		if p.AppId == appID {
			d.Set("version", p.PackageInformation.PackageDefinition.Version)
			d.Set("config", p.PackageInformation.PackageDefinition.Config)
			d.Set("name", p.PackageInformation.PackageDefinition.Name)
			d.SetId(appID)
			return nil
		}
	}

	// app_id defined but no package installed
	d.SetId("")
	return nil

	// localVarOptionals := &dcos.ServiceDescribeOpts{}
	// localVarOptionals.CosmosServiceDescribeV1Request.
	//
	// client.Cosmos.ServiceDescribe(ctx, localVarOptionals)

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

	_, resp, err := client.Cosmos.PackageUninstall(ctx, cosmosPackageUninstallV1Request, nil)

	log.Printf("[TRACE] Cosmos.PackageUninstall - %v", resp)

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
