package dcos

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/imdario/mergo"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

type packageConfigSpec struct {
	Version *packageVersionSpec    `json:"v,omitempty"`
	Config  map[string]interface{} `json:"c"`

	// The `Checksum` is used to compute a unique ID for this data resource,
	// that can be later used by the `dcos_package` resource to consider
	// package re-deployment even if the configuration map has not changed.
	//
	// For example, if the package configuration refers to a secret and the value
	// of that secret has changed (even though it's name is intact), the `Config`
	// map would be identical, by the `Checksum` will differ.
	//
	Checksum string `json:"s"`
}

func dataSourceDcosPackageConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosPackageConfigRead,
		Schema: map[string]*schema.Schema{
			"version_spec": schemaInPackageVersionSpec(false),
			"extend":       schemaInPackageConfigSpec(false),
			"config":       schemaOutPackageConfigSpec(),
			"autotype": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "If `true`, the provider will try to convert string JSON values to their respective types",
			},
			"checksum": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Arbitrary string segments that can be used to calculate a unique configuration checksum",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

/**
 * schemaOutPackageConfigSpec returns a re-usable schema definition that other
 * resources can use as a package config output.
 */
func schemaOutPackageConfigSpec() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "The package configuration specifications",
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
				"config": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"csum": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

/**
 * schemaInPackageConfigSpec returns a re-usable schema definition that other
 * resources can use as a package config input.
 */
func schemaInPackageConfigSpec(required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Required:    required,
		Optional:    !required,
		Description: "The package configuration specifications",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"package": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"version": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"schema": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
				},
				"config": {
					Type:     schema.TypeString,
					Required: true,
					Default:  "",
				},
				"csum": {
					Type:     schema.TypeString,
					Required: true,
					Default:  "",
				},
			},
		},
	}
}

func serializePackageConfigSpec(model *packageConfigSpec) (map[string]interface{}, error) {
	var err error

	// Serialize the package configuration as a JSON
	bSpec, err := json.Marshal(model.Config)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]interface{})
	ret["config"] = string(bSpec)

	// If we have version spec, flat-merge the fields from the serialized
	// version specifications in the package config spec (terraform does not
	// support nested maps)
	if model.Version != nil {
		ver, err := serializePackageVersionSpec(model.Version)
		if err != nil {
			return nil, fmt.Errorf("Unable to serialize version spec: %s", err.Error())
		}

		err = mergo.MergeWithOverwrite(&ret, &ver)
		if err != nil {
			return nil, fmt.Errorf("Could not merge config and version spec")
		}
	}

	ret["csum"] = model.Checksum
	return ret, nil
}

func deserializePackageConfigSpec(spec map[string]interface{}) (*packageConfigSpec, error) {
	var resp packageConfigSpec
	var err error

	// Parse the package configuration JSON
	if v, ok := spec["config"]; ok {
		if s, ok := v.(string); ok {
			err = json.Unmarshal([]byte(s), &resp.Config)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse configuration spec '%s': %s", s, err.Error())
			}
		} else {
			return nil, fmt.Errorf("Field `config` is not string")
		}
	} else {
		return nil, fmt.Errorf("Field `config` is missing")
	}

	// All three fields `package`, `version` and `spec` are coming flat from the
	// version specifications. Since all 3 are normally required, it's enough to
	// just test if either of them is given.
	if v, ok := spec["package"]; ok {
		if s, ok := v.(string); ok && s != "" {
			resp.Version, err = deserializePackageVersionSpec(spec)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse version spec: %s", err.Error())
			}
		} else {
			return nil, fmt.Errorf("Field `package` is not a string")
		}
	} else {
		return nil, fmt.Errorf("Field `package` is missing")
	}

	// Extract the package checksum
	if v, ok := spec["csum"]; ok {
		if s, ok := v.(string); ok {
			resp.Checksum = s
		} else {
			return nil, fmt.Errorf("Field `csum` is not string")
		}
	} else {
		return nil, fmt.Errorf("Field `csum` is missing")
	}

	return &resp, nil
}

/**
 * computeCsum calculates a SHA256 checksum out of the given list of strings
 */
func computeCsum(previous string, data []interface{}) string {
	var strList []string
	strList = append(strList, previous)
	for _, s := range data {
		strList = append(strList, s.(string))
	}
	sum := sha256.Sum256([]byte(strings.Join(strList, "\n")))
	return fmt.Sprintf("%x", sum)
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
					log.Printf("[TRACE] Matched bool: '%t'", tBool)
					ptr[targetKey] = tBool
				} else if json.Unmarshal(data, &tString) == nil {
					log.Printf("[TRACE] Matched string: '%s'", tString)
					ptr[targetKey] = tString
				} else if json.Unmarshal(data, &tList) == nil {
					log.Printf("[TRACE] Matched list: '%v'", tList)
					ptr[targetKey] = tList
				} else if json.Unmarshal(data, &tMap) == nil {
					log.Printf("[TRACE] Matched map: '%v'", tMap)
					ptr[targetKey] = tMap
				} else {
					return nil, fmt.Errorf("Invalid JSON contents encountered")
				}

				log.Printf("[TRACE] Target ptr is now: '%s'", ptr)

			} else {
				return nil, fmt.Errorf("Expecting `json` to be string")
			}
		} else if vMap != nil {
			log.Printf("[TRACE] Processing string/string map: %s", vMap)

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
			log.Printf("[TRACE] Processing string list: %s", vList)

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

		log.Printf("[TRACE] Resulted to: %v", recMap)

		err = mergo.MergeWithOverwrite(&ret, &recMap)
		if err != nil {
			return nil, fmt.Errorf("Could not merge section %d: %s", idx, err.Error())
		}

		log.Printf("[TRACE] Merged to: %v", ret)
	}

	return ret, nil
}

func dataSourceDcosPackageConfigRead(d *schema.ResourceData, meta interface{}) error {
	var configSpec *packageConfigSpec = nil
	var err error

	// Start with a previous configuration spec. Either from the one received
	// from the upstrea, or from a blank slate.
	fromSpec := d.Get("extend").(map[string]interface{})
	if len(fromSpec) != 0 {
		log.Printf("[INFO] Parsing previous config spec: %v", fromSpec)
		configSpec, err = deserializePackageConfigSpec(fromSpec)
		if err != nil {
			return fmt.Errorf("Unable to process `extend` contents: %s", err.Error())
		}
		log.Printf("[INFO] Parsed previous config spec to: %v", configSpec)
	} else {
		log.Printf("[INFO] Not using previous config spec")
		configSpec = &packageConfigSpec{
			nil, make(map[string]interface{}), "",
		}
	}

	versionSpec := d.Get("version_spec").(map[string]interface{})
	if len(versionSpec) != 0 {
		log.Printf("[INFO] Parsing version spec: %v", versionSpec)
		spec, err := deserializePackageVersionSpec(versionSpec)
		if err != nil {
			return fmt.Errorf("Unable to process `version_spec` contents: %s", err.Error())
		}
		log.Printf("[INFO] Parsed version spec to: %v", spec)
		configSpec.Version = spec
	} else {
		log.Printf("[INFO] Configuration block does not include version spec")
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
	log.Printf("[TRACE] User config merged to: %s", util.PrintJSON(&configSpec.Config))

	// Compute a unique checksum from the checksum string fields
	configSpec.Checksum = computeCsum(
		configSpec.Checksum,
		d.Get("checksum").([]interface{}),
	)
	log.Printf("[TRACE] Computing checksum of %v to: %s", d.Get("checksum"), configSpec.Checksum)

	// Serialize config spec into a map
	configMap, err := serializePackageConfigSpec(configSpec)
	if err != nil {
		return fmt.Errorf("Unable to serialize the package config spec: %s", err.Error())
	}
	d.Set("config", configMap)

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
