package dcos

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/antihax/optional"
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosPackageRepo() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosPackageRepoCreate,
		Read:   resourceDcosPackageRepoRead,
		Delete: resourceDcosPackageRepoDelete,

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "Universe",
				Description: "The name of the repository",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "https://universe.mesosphere.com/repo",
				Description: "The URL of the repository",
			},
			"volatile": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "If set to `true`, the repository will be deleted when the resource is un-installed",
			},
			"index": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Default:     -1,
				Description: "Defines the index where this repository will be installed at.",
			},
		},
	}
}

func resourceDcosPackageRepoCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()
	log.Println("[DEBUG] Creating package repository")

	index := d.Get("index").(int)
	repoName := d.Get("name").(string)
	repoUrl := d.Get("url").(string)
	repoAddRequest := dcos.CosmosPackageAddRepoV1Request{
		Name: repoName,
		Uri:  repoUrl,
	}

	if index >= 0 {
		idx := int32(index)
		repoAddRequest.Index = &idx
	}

	_, _, err := client.Cosmos.PackageRepositoryAdd(ctx, &dcos.PackageRepositoryAddOpts{
		CosmosPackageAddRepoV1Request: optional.NewInterface(repoAddRequest),
	})
	if err != nil {
		// Properly extract the underlying CosmosError in order to find out
		// critical cases (such as case where the repository already exists)
		if oaErr, ok := err.(dcos.GenericOpenAPIError); ok {
			if oaErr.Model() != nil {
				if cosmosErr, ok := oaErr.Model().(dcos.CosmosError); ok {
					if cosmosErr.Type == "RepositoryAlreadyPresent" {
						log.Printf("[DEBUG] A repository with the same name/url is already present: %s\n", cosmosErr.Message)
						d.SetId(fmt.Sprintf("%s:%s", repoName, repoUrl))
						return nil
					} else {
						return fmt.Errorf("Error adding repository: %s error: %s", cosmosErr.Type, cosmosErr.Message)
					}
				}
			}
		}

		return fmt.Errorf("Unable to place a repository add request: %s", err.Error())
	}

	// As the "ID" we are using the name/URL combo, separated with a character
	// that cannot appear neither in the URL nor the name ':'
	d.SetId(fmt.Sprintf("%s:%s", repoName, repoUrl))

	return resourceDcosPackageRepoRead(d, meta)
}

func resourceDcosPackageRepoRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()
	log.Printf("[DEBUG] Reading package repository: id='%s'", d.Id())

	resp, _, err := client.Cosmos.PackageRepositoryList(ctx, make(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("Unable to enumerate repositories: %s", err.Error())
	}

	// Separate Name/URL from the ID
	nameUri := strings.SplitN(d.Id(), ":", 2)
	for _, repo := range resp.Repositories {
		if repo.Name == nameUri[0] || repo.Uri == nameUri[1] {
			// If for any reason we have partial match (only name or only URL)
			// then a create will definitely fail. Remove the resource
			if !(repo.Name == nameUri[0] && repo.Uri == nameUri[1]) {
				log.Printf(
					"[WARN] Mismatched requested (name='%s', url='%s') and existing (name='%s', url='%s') repositories",
					nameUri[0],
					nameUri[1],
					repo.Name,
					repo.Uri,
				)
				d.SetId("")
				return nil
			}

			// Otherwise we just found our match
			d.Set("name", repo.Name)
			d.Set("url", repo.Uri)
			return nil
		}
	}

	// We are intentionally not reading the `index` property because
	// it is only used as hinting during creation.

	// Otherwise such repo was not found
	d.SetId("")
	return nil
}

func resourceDcosPackageRepoDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()
	log.Println("[DEBUG] Deleting package repository")

	// Separate Name/URL from the ID
	nameUri := strings.SplitN(d.Id(), ":", 2)
	repoDelRequest := dcos.CosmosPackageDeleteRepoV1Request{
		Name: nameUri[0],
		Uri:  nameUri[1],
	}

	// Place delete request only if the resource is volatile
	if d.Get("volatile").(bool) {
		_, _, err := client.Cosmos.PackageRepositoryDelete(ctx, &dcos.PackageRepositoryDeleteOpts{
			CosmosPackageDeleteRepoV1Request: optional.NewInterface(repoDelRequest),
		})
		if err != nil {
			log.Printf("[DEBUG] Encountered error: %s", err)
			if oaErr, ok := err.(dcos.GenericOpenAPIError); ok {
				log.Printf("[DEBUG] Encountered GenericOpenAPIError: %s", oaErr)
				if oaErr.Model() != nil {
					log.Printf("[DEBUG] Found model: %s", oaErr.Model())
					if cosmosErr, ok := oaErr.Model().(dcos.CosmosError); ok {
						return fmt.Errorf("Error deleting repository: %s error: %s", cosmosErr.Type, cosmosErr.Message)
					}
				}
			}
			return fmt.Errorf("Unable to delete the repository: %s", err.Error())
		}

	}

	d.SetId("")
	return nil
}
