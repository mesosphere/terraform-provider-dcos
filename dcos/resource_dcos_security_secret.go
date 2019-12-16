package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosSecuritySecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosSecuritySecretCreate,
		Read:   resourceDcosSecuritySecretRead,
		Update: resourceDcosSecuritySecretUpdate,
		Delete: resourceDcosSecuritySecretDelete,
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
			"path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  false,
				Sensitive: true,
			},
			"store": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
			},
		},
	}
}

func encodePath(pathToSecret string) string {
	return url.PathEscape(pathToSecret)
}

func resourceDcosSecuritySecretCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	secretsV1Secret := dcos.SecretsV1Secret{}
	secretsV1Secret.Value = d.Get("value").(string)

	pathToSecret := d.Get("path").(string)

	store := d.Get("store").(string)

	// Try to create the secret on DC/OS
	resp, err := client.Secrets.CreateSecret(ctx, store, encodePath(pathToSecret), secretsV1Secret)
	log.Printf("[TRACE] Create %s, %s - %v", store, pathToSecret, resp)
	if err != nil {
		// If this was a conflict, replace the secret
		if strings.Contains(err.Error(), "Conflict") {
			resp, err := client.Secrets.UpdateSecret(ctx, store, encodePath(pathToSecret), secretsV1Secret)
			log.Printf("[TRACE] Update %s, %s - %v", store, pathToSecret, resp)
			if err != nil {
				return fmt.Errorf("Unable to update existing secret: %s", err.Error())
			}
		} else {
			return fmt.Errorf("Unable to create secret: %s", err.Error())
		}
	}

	d.SetId(generateID(store, pathToSecret))
	return nil
}

func resourceDcosSecuritySecretRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	store := d.Get("store").(string)
	pathToSecret := d.Get("path").(string)

	secret, resp, err := client.Secrets.GetSecret(ctx, store, encodePath(pathToSecret), nil)

	log.Printf("[TRACE] Read - %v", resp)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] Read - %s not found", pathToSecret)
		d.SetId("")
		return nil
	}

	if err != nil {
		return nil
	}

	d.Set("value", secret.Value)
	d.SetId(generateID(store, pathToSecret))

	return nil
}

func generateID(store string, pathToSecret string) string {
	return path.Join(store, pathToSecret)
}

func resourceDcosSecuritySecretUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	secretsV1Secret := dcos.SecretsV1Secret{}
	secretsV1Secret.Value = d.Get("value").(string)

	pathToSecret := d.Get("path").(string)

	store := d.Get("store").(string)

	_, err := client.Secrets.UpdateSecret(ctx, store, encodePath(pathToSecret), secretsV1Secret)

	if err != nil {
		return fmt.Errorf("Unable to update secret: %s", err.Error())
	}

	return resourceDcosSecuritySecretRead(d, meta)
}

func resourceDcosSecuritySecretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	pathToSecret := d.Get("path").(string)
	store := d.Get("store").(string)

	resp, err := client.Secrets.DeleteSecret(ctx, store, pathToSecret)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Unable to delete secret: %s", err.Error())
	}

	d.SetId("")
	return nil
}
