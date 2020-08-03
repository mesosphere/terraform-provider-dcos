package dcos

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDcosServiceHttpRequest() *schema.Resource {
	return &schema.Resource{
		Create: resourceDcosServiceHttpRequestCreate,
		Read:   resourceDcosServiceHttpRequestRead,
		Delete: resourceDcosServiceHttpRequestDelete,

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the service on DC/OS.",
			},
			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "/",
				Description: "The path within the service URL.",
			},
			"run_on": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "create",
				Description: "At which lifetime of this resource to make the request.",
			},
			"method": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "GET",
				Description: "The method of the HTTP request to place.",
			},
			"header": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional HTTP headers to include",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"body": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "The raw body to include in the request.",
			},
		},
	}
}

func placeRequest(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(*dcos.APIClient)
	config := apiClient.CurrentDCOSConfig()

	cfgHeaders := d.Get("header").([]interface{})
	cfgServiceName := d.Get("service_name").(string)
	cfgPath := d.Get("path").(string)
	cfgBody := []byte(d.Get("body").(string))
	cfgMethod := d.Get("method").(string)

	url := fmt.Sprintf("%s/service/%s%s", config.URL(), cfgServiceName, cfgPath)

	log.Printf("[TRACE] Posting body: %s", cfgBody)

	request, err := http.NewRequest(cfgMethod, url, bytes.NewReader(cfgBody))
	if err != nil {
		return fmt.Errorf("Unable to prepare request: %s", err.Error())
	}

	request.Header.Add("Authorization", config.ACSToken())
	for _, hdr := range cfgHeaders {
		if recMap, ok := hdr.(map[string]interface{}); ok {
			if iName, ok := recMap["name"]; ok {
				if name, ok := iName.(string); ok {
					if iValue, ok := recMap["value"]; ok {
						if value, ok := iValue.(string); ok {
							request.Header.Add(name, value)
						}
					}
				}
			}
		}
	}

	response, err := apiClient.HTTPClient().Do(request)
	if err != nil {
		return fmt.Errorf("Unable to place request: %s", err.Error())
	}
	defer response.Body.Close()

	log.Printf("[TRACE] Server responded with %s", response.Status)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Server on %s responded with %s", url, response.Status)
	}

	return nil
}

func resourceDcosServiceHttpRequestCreate(d *schema.ResourceData, meta interface{}) error {
	runTrigger := d.Get("run_on").(string)

	if runTrigger == "create" {
		err := placeRequest(d, meta)
		if err != nil {
			return err
		}
	}

	d.SetId("ok")

	return nil
}

func resourceDcosServiceHttpRequestRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDcosServiceHttpRequestDelete(d *schema.ResourceData, meta interface{}) error {
	runTrigger := d.Get("run_on").(string)

	if runTrigger == "delete" {
		err := placeRequest(d, meta)
		if err != nil {
			return err
		}
	}

	d.SetId("")
	return nil
}
