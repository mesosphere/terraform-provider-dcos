package dcos

import (
	"context"
	"fmt"
	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"dcos_acs_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The DC/OS access token",
			},
			"ssl_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     "",
				Description: "Verify SSL connection",
			},
			"dcos_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "URL of DC/OS to use",
			},
			"cluster": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Clustername to use",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "User name logging into the cluster",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Password to login with",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			// "dcos_services_single_container": resourceDcosServicesSingleContainer(),
			// "dcos_job":                       resourceDcosJob(),
			"dcos_secret":              resourceDcosSecret(),
			"dcos_iam_service_account": resourceDcosIAMServiceAccount(),
			"dcos_iam_saml_provider":   resourceDcosSAMLProvider(),
			"dcos_iam_grant_user":      resourceDcosIAMGrantUser(),
			"dcos_package":             resourceDcosPackage(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dcos_service": dataSourceDcosService(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	var config *dcos.Config
	var iamLogin dcos.IamLoginObject
	var login bool = false
	var err error

	// Configure custom cluster URL
	dcosUrl := d.Get("dcos_url").(string)
	if dcosUrl != "" {
		config = dcos.NewConfig(nil)
		config.SetURL(dcosUrl)

		// Require a log-in username
		iamLogin.Uid = d.Get("user").(string)
		if iamLogin.Uid == "" {
			return nil, fmt.Errorf("Missing required 'user' field")
		}

		// Populate the IAM Login object based on the arguments given
		dcosACSToken := d.Get("dcos_acs_token").(string)
		if dcosACSToken == "" {
			loginPass := d.Get("password").(string)
			if loginPass == "" {
				return nil, fmt.Errorf("You must either provide a 'dcos_acs_token' or a 'password' field")
			}
			iamLogin.Password = loginPass
		} else {
			iamLogin.Token = dcosACSToken
		}

		// Login after we have the client
		login = true
	} else {
		// Get current config
		config, err = dcos.NewConfigManager(nil).Current()
		if err != nil {
			return nil, fmt.Errorf("Unable to get default configuration: %s", err.Error())
		}
	}

	// Disable TLS verify if required
	tlsVerify := d.Get("ssl_verify").(bool)
	if !tlsVerify {
		tls := config.TLS()
		tls.Insecure = true
		config.SetTLS(tls)
	}

	// Change the name of the cluster if requested
	clusterName := d.Get("cluster").(string)
	if clusterName != "" {
		config.SetName(clusterName)
	}

	// Create a new DC/OS client
	client, err := dcos.NewClientWithConfig(config)
	if err != nil {
		return nil, err
	}

	if login {
		// Login and obtain an ACS token
		authToken, _, err := client.IAM.Login(context.TODO(), iamLogin)
		if err != nil {
			return nil, fmt.Errorf("Unable to authenticate: %s", err.Error())
		}

		// Update configuration object (it's passed by reference)
		config.SetACSToken(authToken.Token)
	}

	return client, err
}
