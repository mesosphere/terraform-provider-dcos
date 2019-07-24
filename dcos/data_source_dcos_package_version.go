package dcos

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mesosphere-incubator/cosmos-repo-go/cosmos"
)

type packageVersionSpec struct {
	Name    string                 `json:"n"`
	Version string                 `json:"v"`
	Schema  map[string]interface{} `json:"c"`
}

func dataSourceDcosPackageVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosPackageVersionRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the package to install",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "latest",
				Description: "The version of the package to install",
			},
			"repo_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://universe.mesosphere.com/repo",
				Description: "The repository URL to use for resolving the package configuration",
			},
			"spec": schemaOutPackageVersionSpec(),
		},
	}
}

func schemaOutPackageVersionSpec() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "The package version specifications",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"package": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"version": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"schema": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func schemaInPackageVersionSpec(required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Required:    required,
		Optional:    !required,
		Description: "The package version specifications",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"package": {
					Type:     schema.TypeString,
					Required: true,
				},
				"version": {
					Type:     schema.TypeString,
					Required: true,
				},
				"schema": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func serializeCosmosPackage(pkg cosmos.CosmosPackage) (map[string]interface{}, error) {
	return serializePackageVersionSpec(&packageVersionSpec{
		Name:    pkg.GetName(),
		Version: pkg.GetVersion(),
		Schema:  pkg.GetConfig(),
	})
}

func serializePackageVersionSpec(pkg *packageVersionSpec) (map[string]interface{}, error) {
	bSpec, err := json.Marshal(pkg.Schema)
	if err != nil {
		return nil, err
	}

	var gzBytesBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBytesBuf)
	if _, err := gz.Write(bSpec); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	ret := make(map[string]interface{})
	ret["package"] = pkg.Name
	ret["version"] = pkg.Version
	ret["schema"] = base64.StdEncoding.EncodeToString(gzBytesBuf.Bytes())
	return ret, nil
}

func deserializePackageVersionSpec(spec map[string]interface{}) (*packageVersionSpec, error) {
	var resp packageVersionSpec

	if v, ok := spec["package"]; ok {
		if s, ok := v.(string); ok {
			resp.Name = s
		} else {
			return nil, fmt.Errorf("Field `name` is not string")
		}
	} else {
		return nil, fmt.Errorf("Field `name` is missing")
	}

	if v, ok := spec["version"]; ok {
		if s, ok := v.(string); ok {
			resp.Version = s
		} else {
			return nil, fmt.Errorf("Field `version` is not string")
		}
	} else {
		return nil, fmt.Errorf("Field `version` is missing")
	}

	if v, ok := spec["schema"]; ok {
		if s, ok := v.(string); ok {
			gzBytes, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, fmt.Errorf("Unable to decode spec: %s", err.Error())
			}

			zReader, err := gzip.NewReader(bytes.NewReader(gzBytes))
			if err != nil {
				return nil, fmt.Errorf("Unable to unzip the gzip stream: %s", err.Error())
			}

			bSpec, err := ioutil.ReadAll(zReader)
			if err != nil {
				return nil, fmt.Errorf("Unable to read from gzip stream: %s", err.Error())
			}

			err = json.Unmarshal(bSpec, &resp.Schema)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse schema '%s': %s", s, err.Error())
			}
		} else {
			return nil, fmt.Errorf("Field `schema` is not string")
		}
	} else {
		return nil, fmt.Errorf("Field `schema` is missing")
	}

	return &resp, nil
}

func dataSourceDcosPackageVersionRead(d *schema.ResourceData, meta interface{}) error {
	var pkg cosmos.CosmosPackage
	var err error

	repoUrl := d.Get("repo_url").(string)
	log.Printf("[DEBUG] Downloading repository data from %s", repoUrl)
	repo, err := cosmos.NewRepoFromURL(repoUrl)
	if err != nil {
		return fmt.Errorf("Error loading repository '%s' data: %s", repoUrl, err.Error())
	}

	packageName := d.Get("name").(string)
	packageVersion := d.Get("version").(string)

	if packageVersion == "latest" {
		// Look-up the latest package version if the user specified 'latest'
		pkg, err = repo.FindLatestPackageVersion(packageName)
		if err != nil {
			return fmt.Errorf("Unable to find the latest version of package '%s': %s", packageName, err.Error())
		}
		log.Printf("[DEBUG] Found latest version %s", pkg.GetVersion())
	} else {
		// Otherwise make sure that the package exists in the specified repository
		pkg, err = repo.FindPackageVersion(packageName, packageVersion)
		if err != nil {
			return fmt.Errorf("Unable to find the package '%s' version '%s' in the repository: %s", packageName, packageVersion, err.Error())
		}
	}

	spec, err := serializeCosmosPackage(pkg)
	if err != nil {
		return fmt.Errorf("Unable to marshal package specifications: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", packageName, packageVersion))
	d.Set("spec", spec)

	return nil
}
