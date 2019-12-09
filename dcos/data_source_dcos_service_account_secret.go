package dcos

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

type ServiceAccontJson struct {
	Scheme        string `json:"scheme"`
	UID           string `json:"uid"`
	LoginEndpoint string `json:"login_endpoint"`
	PrivateKey    string `json:"private_key"`
}

func dataSourceDcosServiceAccountSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosServiceAccountSecretRead,
		Schema: map[string]*schema.Schema{
			"private_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The PEM encoded PKCS1 (RSA) or PKCS8 private key for this account",
			},
			"uid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user ID",
			},
			"login_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "https://leader.mesos/acs/api/v1/auth/login",
				Description: "Customize the login_endpoint parameter",
			},
			"contents": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The JSON contents for the secret to upload",
			},
		},
	}
}

func getKeyId(pubKey crypto.PublicKey) (string, error) {
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	hasher := crypto.SHA256.New()
	_, err = hasher.Write(derBytes)
	if err != nil {
		return "", err
	}

	b := hasher.Sum(nil)[:30]
	keyId := strings.TrimRight(base64.StdEncoding.EncodeToString(b), "=")
	return keyId, nil
}

func dataSourceDcosServiceAccountSecretRead(d *schema.ResourceData, meta interface{}) error {
	var (
		pKey   *rsa.PrivateKey = nil
		pemOut strings.Builder
		sa     ServiceAccontJson
		ok     bool
		err    error
	)

	privateKeyContents := d.Get("private_key").(string)
	loginEndpoint := d.Get("login_endpoint").(string)
	uid := d.Get("uid").(string)

	// Try parsing block as PEM
	block, _ := pem.Decode([]byte(privateKeyContents))
	if block == nil {
		return fmt.Errorf("Unable to decode private key PEM data")
	}

	// First try parsing it as PKCS8-encapsulated private key
	parseResult, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// If this fails, fall-back into parsing it as a plain PKCS1 private key
		pKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("Unable to parse PKCS8 or PKCS1 private key")
		}
	} else {
		pKey, ok = parseResult.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("Unable to parse RSA private key")
		}
	}

	err = pKey.Validate()
	if err != nil {
		return fmt.Errorf("Private key is not valid")
	}

	// Re-encode as PKCS8 private key
	keyBytes, err := x509.MarshalPKCS8PrivateKey(pKey)
	if err != nil {
		return fmt.Errorf("Unable to encode PKCS8 private key")
	}

	// Re-encode as PEM private key
	block = &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   keyBytes,
	}
	err = pem.Encode(&pemOut, block)
	if err != nil {
		return fmt.Errorf("Unable to encode private key PEM data")
	}

	// Compute the key ID
	keyId, err := getKeyId(pKey.Public())
	if err != nil {
		return fmt.Errorf("Unable to compute the key ID")
	}

	// Compose the key
	sa.Scheme = "RS256"
	sa.UID = fmt.Sprintf("")
	sa.LoginEndpoint = loginEndpoint
	sa.PrivateKey = pemOut.String()

	// Compose the result JSON
	retJson, err := json.Marshal(sa)
	if err != nil {
		return fmt.Errorf("Unable to compose service account contents")
	}

	d.Set("contents", string(retJson))
	d.SetId(fmt.Sprintf("%s:%s", uid, keyId))

	return nil
}
