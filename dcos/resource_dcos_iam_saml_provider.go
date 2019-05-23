package dcos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosSAMLProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosSAMLProviderCreate,
		Read:   resourceDcosSAMLProviderRead,
		Update: resourceDcosSAMLProviderUpdate,
		Delete: resourceDcosSAMLProviderDelete,
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Unique Identifier for this Provider. Only lowercase characters allowed",
				ValidateFunc: validateProviderID,
			},
			"idp_metadata": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "IDP Metadata",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if strings.TrimSpace(old) == strings.TrimSpace(new) {
						return true
					}
					return false
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Description of the newly created service account",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Service provider base URL",
			},
			"callback_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SAML Callbackurl",
			},
			"metadata": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SAML service provider metadata",
			},
			"entity_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Provided entity ID",
			},
		},
	}
}

func validateProviderID(i interface{}, k string) (s []string, es []error) {
	provider_id := i.(string)

	if provider_id != strings.ToLower(provider_id) {
		es = append(es, fmt.Errorf("provider_id %s contains uppercase characters. Only lowercase allowed", provider_id))
		return
	}

	return
}

func resourceDcosSAMLProviderCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	providerId := d.Get("provider_id").(string)
	idpMetadata := d.Get("idp_metadata").(string)
	spBaseURL := d.Get("base_url").(string)

	var iamsamlProviderConfig dcos.IamsamlProviderConfig

	if description, ok := d.GetOk("description"); ok {
		iamsamlProviderConfig.Description = description.(string)
	}

	iamsamlProviderConfig.IdpMetadata = idpMetadata
	iamsamlProviderConfig.SpBaseUrl = spBaseURL

	resp, err := client.IAM.ConfigureSAMLProvider(ctx, providerId, iamsamlProviderConfig)
	log.Printf("[TRACE] IAM.ConfigureSAMLProvider - %v", resp)

	if err != nil {
		if iamErr, ok := iamErrorOK(err); ok {
			return iamPrettyError(iamErr)
		}
		return err
	}

	return resourceDcosSAMLProviderRead(d, meta)
}

func resourceDcosSAMLProviderRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	providerId := d.Get("provider_id").(string)
	providerConfig, resp, err := client.IAM.GetSAMLProvider(ctx, providerId)

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if err != nil {
		if iamErr, ok := iamErrorOK(err); ok {
			return iamPrettyError(iamErr)
		}
		return err
	}

	d.Set("description", providerConfig.Description)
	d.Set("idp_metadata", strings.TrimSpace(providerConfig.IdpMetadata))
	d.Set("base_url", providerConfig.SpBaseUrl)

	if callbackurl, _, err := client.IAM.GetSAMLProviderACSCallbackURL(ctx, providerId); err == nil {
		d.Set("callback_url", callbackurl.AcsCallbackUrl)
	} else {
		d.Set("callback_url", "")
	}

	metadata, metadataResp, err := client.IAM.GetSAMLProviderSPMetadata(ctx, providerId)
	dumpReq, _ := httputil.DumpRequest(metadataResp.Request, false)
	log.Printf("[TRACE] IAM.GetSAMLProviderSPMetadata - %v", metadataResp)
	log.Printf("[TRACE] IAM.GetSAMLProviderSPMetadata - Request %s", dumpReq)

	if err != nil {
		log.Printf("[WARNING] IAM.GetSAMLProviderSPMetadata Error - %v", err)
		d.Set("metadata", "")
		d.Set("entity_id", "")
	} else {
		d.Set("metadata", metadata)

		doc := etree.NewDocument()

		if err := doc.ReadFromString(metadata); err == nil {
			if element := doc.SelectElement("md:EntityDescriptor"); element != nil {
				if entityID := element.SelectAttr("entityID"); entityID != nil {
					d.Set("entity_id", entityID.Value)
				}
			}
		} else {
			log.Printf("[WARNING] IAM.GetSAMLProviderSPMetadata XML Error - %v", err)
			d.Set("entity_id", "")
		}
	}

	d.SetId("provider_id")

	return nil
}

func resourceDcosSAMLProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	var iamsamlProviderConfig dcos.IamsamlProviderConfig

	providerId := d.Get("provider_id").(string)
	idpMetadata := d.Get("idp_metadata").(string)
	spBaseURL := d.Get("base_url").(string)

	if description, ok := d.GetOk("description"); ok {
		iamsamlProviderConfig.Description = description.(string)
	}

	iamsamlProviderConfig.IdpMetadata = idpMetadata
	iamsamlProviderConfig.SpBaseUrl = spBaseURL

	resp, err := client.IAM.UpdateSAMLProvider(ctx, providerId, iamsamlProviderConfig)

	log.Printf("[TRACE] IAM.UpdateSAMLProvider - %v", resp)

	if err != nil {
		if iamErr, ok := iamErrorOK(err); ok {
			return iamPrettyError(iamErr)
		}
		return err
	}

	return resourceDcosSAMLProviderRead(d, meta)
}

func resourceDcosSAMLProviderDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*dcos.APIClient)
	ctx := context.TODO()

	providerId := d.Get("provider_id").(string)

	resp, err := client.IAM.DeleteSAMLProvider(ctx, providerId)

	log.Printf("[TRACE] IAM.DeleteSAMLProvider - %v", resp)

	if err != nil {
		if iamErr, ok := iamErrorOK(err); ok {
			return iamPrettyError(iamErr)
		}
		return err
	}

	d.SetId("")

	return nil
}
