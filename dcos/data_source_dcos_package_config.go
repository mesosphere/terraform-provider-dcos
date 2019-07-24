package dcos

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/imdario/mergo"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

type packageConfigSpec struct {
	Version *packageVersionSpec    `json:"v,omitempty"`
	Config  map[string]interface{} `json:"c"`
}

func dataSourceDcosPackageConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosPackageConfigRead,
		Schema: map[string]*schema.Schema{
			"version_spec": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The package version to install",
			},
			"extend": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The previous package configuration to extend",
			},
			"config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration output (can be chained to other config's `extend`)",
			},
			"autotype": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "If `true`, the provider will try to convert string JSON values to their respective types",
			},
			"section": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Description: "Environment variables",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"json": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"section.map", "section.list"},
						},
						"list": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:      true,
							ConflictsWith: []string{"section.json", "section.map"},
						},
						"map": {
							Type:          schema.TypeMap,
							Elem:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"section.json", "section.list"},
						},
					},
				},
			},
		},
	}
}

func serializePackageConfigSpec(model *packageConfigSpec) (string, error) {
	bSpec, err := json.Marshal(model)
	if err != nil {
		return "", err
	}

	var gzBytesBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBytesBuf)
	if _, err := gz.Write(bSpec); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzBytesBuf.Bytes()), nil
}

func deserializePackageConfigSpec(spec string) (*packageConfigSpec, error) {
	var resp *packageConfigSpec

	gzBytes, err := base64.StdEncoding.DecodeString(spec)
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

	err = json.Unmarshal(bSpec, &resp)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse configuration spec '%s': %s", spec, err.Error())
	}
	return resp, nil
}

func sectionToJson(section interface{}, autotype bool) (map[string]interface{}, error) {
	var targetKey string
	ret := make(map[string]interface{})

	log.Printf("[TRACE] Processing section: %s", section)
	if recMap, ok := section.(map[string]interface{}); ok {
		path, _ := recMap["path"]
		if path == nil {
			return nil, fmt.Errorf("Missing value for key: path")
		}
		pathStr, ok := path.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid type of key: path")
		}

		log.Printf("[TRACE] Path: %s", pathStr)

		vMap, _ := recMap["map"]
		vList, _ := recMap["list"]
		vJson, _ := recMap["json"]

		// Require one of `map`, `list`, `json`
		if vMap == nil && vList == nil && vJson == "" {
			return nil, fmt.Errorf("Require at least one of `map`, `list` or `json`")
		}

		// Walk down the path up to the one-by-last element of the given path
		// and assign it on the `ptr` variable
		ptr := ret
		walkedPath := ""
		parts := strings.Split(pathStr, ".")
		targetKey, parts = parts[len(parts)-1], parts[:len(parts)-1]
		log.Printf("[TRACE] Broken path '%s' into: '%s' and key '%s", pathStr, parts, targetKey)

		for _, part := range parts {
			log.Printf("[TRACE] Walking part '%s' in: %s", part, ptr)

			// Keep track of the path we have walked so far, for debugging purposes
			if walkedPath != "" {
				walkedPath += "."
			}
			walkedPath += part

			log.Printf("[TRACE] Currently in: '%s'", walkedPath)

			// Make sure that for every part we have a map
			if child, ok := ptr[part]; ok {
				if childMap, ok := child.(map[string]interface{}); ok {
					ptr = childMap
				} else {
					return nil, fmt.Errorf(
						"Did not encounter an object in '%s' while looking for '%s'",
						walkedPath,
						pathStr,
					)
				}
			} else {
				ptr[part] = make(map[string]interface{})
				ptr = ptr[part].(map[string]interface{})
			}
		}

		if vJson != "" {
			if jsonValue, ok := vJson.(string); ok {
				var tString string
				var tNumber float64
				var tBool bool
				var tList []interface{}
				var tMap map[string]interface{}

				log.Printf("[TRACE] Unserializing raw JSON '%s'", jsonValue)

				// Try unmarshalling into various types
				data := []byte(jsonValue)
				if json.Unmarshal(data, &tNumber) == nil {
					log.Printf("[TRACE] Matched number: '%f'", tNumber)
					ptr[targetKey] = tNumber
				} else if json.Unmarshal(data, &tBool) == nil {
					log.Printf("[TRACE] Matched bool: '%s'", tBool)
					ptr[targetKey] = tBool
				} else if json.Unmarshal(data, &tString) == nil {
					log.Printf("[TRACE] Matched string: '%s'", tString)
					ptr[targetKey] = tString
				} else if json.Unmarshal(data, &tList) == nil {
					log.Printf("[TRACE] Matched list: '%s'", tList)
					ptr[targetKey] = tList
				} else if json.Unmarshal(data, &tMap) == nil {
					log.Printf("[TRACE] Matched map: '%s'", tMap)
					ptr[targetKey] = tMap
				} else {
					return nil, fmt.Errorf("Invalid JSON contents encountered")
				}

				log.Printf("[TRACE] Target ptr is now: '%s'", ptr)

			} else {
				return nil, fmt.Errorf("Expecting `json` to be string")
			}
		} else if vMap != nil {
			log.Printf("TRACE] Processing string/string map: %s", vMap)

			if mapValue, ok := vMap.(map[string]interface{}); ok {
				if autotype {
					ptr[targetKey] = util.AutotypeMap(mapValue)
				} else {
					ptr[targetKey] = mapValue
				}
			} else {
				return nil, fmt.Errorf("Expecting `map` to be a map of string/string")
			}
		} else if vList != nil {
			log.Printf("TRACE] Processing string list: %s", vList)

			if listValue, ok := vList.([]interface{}); ok {
				if autotype {
					ptr[targetKey] = util.AutotypeList(listValue)
				} else {
					ptr[targetKey] = listValue
				}
			} else {
				return nil, fmt.Errorf("Expecting `list` to be a list of values")
			}
		}

	} else {
		return nil, fmt.Errorf("Unexpected data type")
	}

	return ret, nil
}

/**
 * Merge individual sections into a continuous JSON object
 */
func mergeSections(sections []interface{}, autotype bool) (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	for idx, rec := range sections {
		log.Printf("[TRACE] Converting section %d to string: %d", idx, rec)
		recMap, err := sectionToJson(rec, autotype)
		if err != nil {
			return nil, fmt.Errorf("On section %d: %s", idx, err.Error())
		}

		log.Printf("[TRACE] Resulted to: %d", recMap)

		err = mergo.MergeWithOverwrite(&ret, &recMap)
		if err != nil {
			return nil, fmt.Errorf("Could not merg section %d: %s", idx, err.Error())
		}

		log.Printf("[TRACE] Merged to: %d", ret)
	}

	return ret, nil
}

func dataSourceDcosPackageConfigRead(d *schema.ResourceData, meta interface{}) error {
	var configSpec *packageConfigSpec = nil
	var err error

	// Start with a previous configuration spec. Either from the one received
	// from the upstrea, or from a blank slate.
	fromSpec := d.Get("extend").(string)
	if fromSpec != "" {
		configSpec, err = deserializePackageConfigSpec(fromSpec)
		if err != nil {
			return fmt.Errorf("Unable to process `extend` contents: %s", err.Error())
		}
	} else {
		configSpec = &packageConfigSpec{
			nil, make(map[string]interface{}),
		}
	}

	versionSpec := d.Get("version_spec").(string)
	if versionSpec != "" {
		spec, err := deserializePackageVersionSpec(versionSpec)
		if err != nil {
			return fmt.Errorf("Unable to process `version_spec` contents: %s", err.Error())
		}
		configSpec.Version = spec
	}

	autotype := d.Get("autotype").(bool)
	sections := d.Get("section").([]interface{})
	config, err := mergeSections(sections, autotype)
	if err != nil {
		return fmt.Errorf("Unable to merge configuration sections: %s", err.Error())
	}

	err = mergo.MergeWithOverwrite(&configSpec.Config, &config)
	if err != nil {
		return fmt.Errorf("Could not merge config with upstream: %s", err.Error())
	}

	configStr, err := serializePackageConfigSpec(configSpec)
	if err != nil {
		return fmt.Errorf("Unable to serialize the package config spec: %s", err.Error())
	}

	d.Set("config", configStr)

	// Compute an ID that consists of the version spec and the config spec

	cfgHash, err := util.HashDict(configSpec.Config)
	if err != nil {
		return fmt.Errorf("Unable to hash the config: %s", err.Error())
	}
	if configSpec.Version != nil {
		schemaHash, err := util.HashDict(configSpec.Version.Schema)
		if err != nil {
			return fmt.Errorf("Unable to hash the version: %s", err.Error())
		}
		d.SetId(fmt.Sprintf("%s:%s:%s-%s",
			configSpec.Version.Name,
			configSpec.Version.Version,
			schemaHash,
			cfgHash,
		))
	} else {
		d.SetId(cfgHash)
	}

	return nil
}
