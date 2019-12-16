package dcos

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/dcos/client-go/dcos"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

func dataSourceDcosCLICommand() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosCLICommandRead,
		Schema: map[string]*schema.Schema{
			"package": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The DC/OS package to obtain the CLI from",
			},

			"cli_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The DC/OS cli to use. If missing defaults to the one from the cluster version.",
			},

			"read":   executeSchema("The command to invoke for reading the resource state"),
			"create": executeSchema("The command to invoke for creating the resource"),
			"update": executeSchema("The command to invoke for updating the resource"),
			"delete": executeSchema("The command to invoke for deleting the resource"),

			"wait_create": executeSchema("The command to invoke for waiting until the resource is created"),
			"wait_update": executeSchema("The command to invoke for waiting until the resource is updated"),
			"wait_delete": executeSchema("The command to invoke for waiting until the resource is deleted"),

			"sandbox": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The sandbox to use for the cli",
				Default:     ".terraform/dcos/sandbox",
			},

			"schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration spec that can be passed to the cli resource.",
			},
		},
	}
}

type cliExecuteMatchSpec struct {
	StdoutMatch string
	StderrMatch string
	ExitCode    int
}

type cliExecuteConfigSpec struct {
	Args        string
	Stdin       string
	Success     cliExecuteMatchSpec
	Failure     cliExecuteMatchSpec
	CreateFiles map[string]string
	OutputFile  string
}

type cliConfigSpec struct {
	Package          string
	SandboxBaseDir   string
	CliVersion       string
	TemplateVarNames []string

	Create *cliExecuteConfigSpec
	Read   *cliExecuteConfigSpec
	Update *cliExecuteConfigSpec
	Delete *cliExecuteConfigSpec

	WaitCreate *cliExecuteConfigSpec
	WaitRead   *cliExecuteConfigSpec
	WaitUpdate *cliExecuteConfigSpec
	WaitDelete *cliExecuteConfigSpec
}

func executeSchema(desc string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Required:    false,
		Description: desc,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"args": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "The arguments to pass to the command, as a plain string",
				},
				"stdin": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "The payload to send to the CLI standard input",
				},
				"files": {
					Type:        schema.TypeList,
					Required:    false,
					Description: "File(s) to create before invoking the command",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the file to create",
							},
							"contents": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "",
							},
							"command_output": {
								Type:        schema.TypeBool,
								Optional:    true,
								Default:     false,
								Description: "Consider this file as the command output",
							},
						},
					},
				},
				"success": &schema.Schema{
					Type:        schema.TypeMap,
					Required:    false,
					Description: "The conditions that should occur to consider the invocation successful",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"stdout_match": {
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "",
								Description: "The regexp to match on the CLI standard output",
							},
							"stderr_match": {
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "",
								Description: "The regexp to match on the CLI standard error",
							},
							"exitcode": {
								Type:        schema.TypeInt,
								Default:     -1,
								Optional:    true,
								Description: "The exit code that should match",
							},
						},
					},
				},
				"failure": &schema.Schema{
					Type:        schema.TypeMap,
					Required:    false,
					Description: "The conditions that should occur to consider the invocation failed",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"stdout_match": {
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "",
								Description: "The regexp to match on the CLI standard output",
							},
							"stderr_match": {
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "",
								Description: "The regexp to match on the CLI standard error",
							},
							"exitcode": {
								Type:        schema.TypeInt,
								Default:     -1,
								Optional:    true,
								Description: "The exit code that should match",
							},
						},
					},
				},
			},
		},
	}
}

/**
 * Find the variables used in the given input and append them on the list
 */
func collectTemplateVars(input string, list []string) []string {
	var ret []string = list
	re := regexp.MustCompile(`{{([\w-]+)}}`)
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		name := match[1]

		found := false
		for _, prev := range list {
			if prev == name {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, name)
		}
	}

	return ret
}

func parseExecConfigSpec(input interface{}, cs *cliConfigSpec) (*cliExecuteConfigSpec, error) {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expecting value to be a map")
	}

	var ret cliExecuteConfigSpec
	if value, ok := inputMap["args"]; ok {
		ret.Args = value.(string)
		cs.TemplateVarNames = collectTemplateVars(ret.Args, cs.TemplateVarNames)
	}
	if value, ok := inputMap["stdin"]; ok {
		ret.Stdin = value.(string)
		cs.TemplateVarNames = collectTemplateVars(ret.Stdin, cs.TemplateVarNames)
	}
	if value, ok := inputMap["success"]; ok {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if value, ok := valueMap["stdout_match"]; ok {
				ret.Success.StdoutMatch = value.(string)
				cs.TemplateVarNames = collectTemplateVars(ret.Success.StdoutMatch, cs.TemplateVarNames)
			}
			if value, ok := valueMap["stderr_match"]; ok {
				ret.Success.StderrMatch = value.(string)
				cs.TemplateVarNames = collectTemplateVars(ret.Success.StderrMatch, cs.TemplateVarNames)
			}
			if value, ok := valueMap["exitcode"]; ok {
				ret.Success.ExitCode = value.(int)
			}
		} else {
			return nil, fmt.Errorf("Expecting a map on `success`")
		}
	}
	if value, ok := inputMap["failure"]; ok {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if value, ok := valueMap["stdout_match"]; ok {
				ret.Failure.StdoutMatch = value.(string)
				cs.TemplateVarNames = collectTemplateVars(ret.Failure.StdoutMatch, cs.TemplateVarNames)
			}
			if value, ok := valueMap["stderr_match"]; ok {
				ret.Failure.StderrMatch = value.(string)
				cs.TemplateVarNames = collectTemplateVars(ret.Failure.StderrMatch, cs.TemplateVarNames)
			}
			if value, ok := valueMap["exitcode"]; ok {
				ret.Failure.ExitCode = value.(int)
			}
		} else {
			return nil, fmt.Errorf("Expecting a map on `failure`")
		}
	}

	ret.CreateFiles = make(map[string]string)
	ret.OutputFile = ""
	if value, ok := inputMap["files"]; ok {
		if valueList, ok := value.([]interface{}); ok {
			for idx, file := range valueList {
				valueMap, ok := file.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("File #%d: Expecting a map", idx)
				}

				fileNameI, ok := valueMap["name"]
				if !ok {
					return nil, fmt.Errorf("File #%d: Missing name", idx)
				}
				fileName := fileNameI.(string)

				if value, ok := valueMap["contents"]; ok {
					valueStr := value.(string)
					ret.CreateFiles[fileName] = valueStr
					cs.TemplateVarNames = collectTemplateVars(valueStr, cs.TemplateVarNames)
				}
				if value, ok := valueMap["output"]; ok {
					if value.(bool) {
						ret.OutputFile = fileName
					}
				}
			}
		} else {
			return nil, fmt.Errorf("Expecting a list on `files`")
		}
	}

	return &ret, nil
}

func parseCliConfigSpec(model *schema.ResourceData) (*cliConfigSpec, error) {
	var ret cliConfigSpec

	ret.Package = model.Get("package").(string)

	if value, ok := model.GetOk("read"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `read`: %s", err.Error())
		}
		ret.Read = execRet
	}
	if value, ok := model.GetOk("create"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `create`: %s", err.Error())
		}
		ret.Create = execRet
	}
	if value, ok := model.GetOk("update"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `update`: %s", err.Error())
		}
		ret.Update = execRet
	}
	if value, ok := model.GetOk("delete"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `delete`: %s", err.Error())
		}
		ret.Delete = execRet
	}

	if value, ok := model.GetOk("wait_create"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `wait_create`: %s", err.Error())
		}
		ret.WaitCreate = execRet
	}
	if value, ok := model.GetOk("wait_update"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `wait_update`: %s", err.Error())
		}
		ret.WaitUpdate = execRet
	}
	if value, ok := model.GetOk("wait_delete"); ok {
		execRet, err := parseExecConfigSpec(value, &ret)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse `wait_delete`: %s", err.Error())
		}
		ret.WaitDelete = execRet
	}

	return &ret, nil
}

func serializeCliConfigSpec(model *cliConfigSpec) (string, error) {
	var err error
	bSpec, err := json.Marshal(model)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bSpec), nil
}

func deserializeCliConfigSpec(input string) (*cliConfigSpec, error) {
	var resp cliConfigSpec
	if input == "" {
		return nil, nil
	}

	bSpec, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode spec '%s': %s", input, err.Error())
	}

	err = json.Unmarshal(bSpec, &resp)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse spec '%s': %s", input, err.Error())
	}

	return &resp, nil
}

func hashConfig(config string) string {
	sum := sha256.Sum256([]byte(config))
	return fmt.Sprintf("%x", sum)
}

func dataSourceDcosCLICommandRead(d *schema.ResourceData, meta interface{}) error {
	var err error
	client := meta.(*dcos.APIClient)
	//ctx := context.TODO()

	// Create cli spec
	spec, err := parseCliConfigSpec(d)
	if err != nil {
		return err
	}

	// If we don't have an explicit cli version, use the DC/OS version
	cliVersion := d.Get("cli_version").(string)
	if cliVersion == "" {
		verSpec, err := util.DCOSGetVersion(client)
		if err != nil {
			return fmt.Errorf("Unable to get the DC/OS version: %s", err.Error())
		}

		cliVersion = verSpec.Version
	}

	// If there is no sandbox configured, use the default sandbox
	sandboxPath := d.Get("sandbox").(string)
	if sandboxPath == "" {
		sandboxPath = fmt.Sprintf(".terraform/dcos/cli-%s", cliVersion)
	}

	// Try preparing the CLI sandbox and squeeze out any bugs we might
	// encounter as early in this process as possible
	cli, err := util.CreateCliWrapper(sandboxPath, client, cliVersion)
	if err != nil {
		return fmt.Errorf("Unable to create cli wrapper: %s", err.Error())
	}
	err = cli.Prepare()
	if err != nil {
		return fmt.Errorf("Unable to prepare the cli wrapper: %s", err.Error())
	}

	// Update with the locally processed meta-data
	spec.SandboxBaseDir = sandboxPath
	spec.CliVersion = cliVersion

	// Serialize configuration into a string
	specStr, err := serializeCliConfigSpec(spec)
	if err != nil {
		return fmt.Errorf("Unable to serialize spec: %s", err.Error())
	}

	d.Set("spec", specStr)
	d.SetId(hashConfig(specStr))

	return nil
}
