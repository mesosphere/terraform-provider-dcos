package dcos

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosSecurityClusterOIDC() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosSecurityClusterOIDCCreate,
		Read:   resourceDcosSecurityClusterOIDCRead,
		Update: resourceDcosSecurityClusterOIDCUpdate,
		Delete: resourceDcosSecurityClusterOIDCDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique Identifier for this Provider. Only lowercase characters allowed",
			},
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "IDP Metadata",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Description of the newly created service account",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service provider base URL",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SAML Callbackurl",
			},
			"issuer": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SAML service provider metadata",
			},
			"ca_certs": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Provided entity ID",
			},
			"verify_server_certificate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Provided entity ID",
			},
		},
	}
}

func resourceDcosSecurityClusterOIDCCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	providerID := d.Get("provider_id").(string)
	baseURL := d.Get("base_url").(string)
	description := d.Get("description").(string)
	issuer := d.Get("issuer").(string)

	caCerts := d.Get("ca_certs").(string)
	verifyServerCertificate := d.Get("verify_server_certificate").(bool)

	clientID := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)

	iamoidcProviderConfig := dcos.IamoidcProviderConfig{
		BaseUrl: baseURL,
	}

	iamoidcProviderConfig.BaseUrl = baseURL

	iamoidcProviderConfig.ClientId = clientID
	iamoidcProviderConfig.ClientSecret = clientSecret
	iamoidcProviderConfig.Description = description
	iamoidcProviderConfig.Issuer = issuer

	iamoidcProviderConfig.CaCerts = caCerts
	iamoidcProviderConfig.VerifyServerCertificate = verifyServerCertificate

	resp, err := client.IAM.ConfigureOIDCProvider(ctx, providerID, iamoidcProviderConfig)
	log.Printf("[TRACE] IAM.ConfigureOIDCProvider - %v", resp)

	if err != nil {
		return err
	}

	return resourceDcosSecurityClusterOIDCRead(d, meta)
}

func resourceDcosSecurityClusterOIDCRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	providerID := d.Get("provider_id").(string)

	providerConfig, resp, err := client.IAM.GetOIDCProvider(ctx, providerID)

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.Set("description", providerConfig.Description)
	d.Set("base_url", providerConfig.BaseUrl)
	d.Set("issuer", providerConfig.Issuer)

	d.Set("client_id", providerConfig.ClientId)
	d.Set("client_secret", providerConfig.ClientSecret)

	d.Set("ca_certs", providerConfig.ClientId)
	d.Set("verify_server_certificate", providerConfig.ClientSecret)

	d.SetId(providerID)

	return nil
}

func resourceDcosSecurityClusterOIDCUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	var iamoidcProviderConfig dcos.IamoidcProviderConfig

	providerID := d.Get("provider_id").(string)

	if description, ok := d.GetOk("description"); ok {
		iamoidcProviderConfig.Description = description.(string)
	}

	if caCerts, ok := d.GetOk("ca_certs"); ok {
		iamoidcProviderConfig.CaCerts = caCerts.(string)
	}

	if verifyCert, ok := d.GetOk("verify_server_certificate"); ok {
		iamoidcProviderConfig.VerifyServerCertificate = verifyCert.(bool)
	}

	resp, err := client.IAM.UpdateOIDCProvider(ctx, providerID, iamoidcProviderConfig)

	log.Printf("[TRACE] IAM.UpdateOIDCProvider - %v", resp)

	if err != nil {
		return err
	}

	return resourceDcosSecurityClusterOIDCRead(d, meta)
}

func resourceDcosSecurityClusterOIDCDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	providerID := d.Get("provider_id").(string)

	resp, err := client.IAM.DeleteOIDCProvider(ctx, providerID)

	log.Printf("[TRACE] IAM.DeleteOIDCProvider - %v", resp)

	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
