package cosmos

import (
	"fmt"
)

type RepoV20Package struct {
	PackagingVersion string
	Name             string
	Version          string
	Description      string
	Config           map[string]interface{}
}

func (p *RepoV20Package) GetName() string {
	return p.Name
}

func (p *RepoV20Package) GetVersion() string {
	return p.Version
}

func (p *RepoV20Package) GetDescription() string {
	return p.Description
}

func (p *RepoV20Package) GetConfig() map[string]interface{} {
	return p.Config
}

func parseV20Package(input map[string]interface{}) (*RepoV20Package, error) {
	pkg := &RepoV20Package{}
	if v, ok := input["packagingVersion"]; ok {
		if vStr, ok := v.(string); ok {
			pkg.PackagingVersion = vStr
		} else {
			return nil, fmt.Errorf("Invalid `packagingVersion` field type")
		}
	} else {
		return nil, fmt.Errorf("Missing `packagingVersion` field")
	}
	if v, ok := input["name"]; ok {
		if vStr, ok := v.(string); ok {
			pkg.Name = vStr
		} else {
			return nil, fmt.Errorf("Invalid `name` field type")
		}
	} else {
		return nil, fmt.Errorf("Missing `name` field")
	}
	if v, ok := input["version"]; ok {
		if vStr, ok := v.(string); ok {
			pkg.Version = vStr
		} else {
			return nil, fmt.Errorf("Invalid `version` field type")
		}
	} else {
		return nil, fmt.Errorf("Missing `version` field")
	}
	if v, ok := input["description"]; ok {
		if vStr, ok := v.(string); ok {
			pkg.Description = vStr
		} else {
			return nil, fmt.Errorf("Invalid `description` field type")
		}
	} else {
		return nil, fmt.Errorf("Missing `description` field")
	}
	if v, ok := input["config"]; ok {
		if vStr, ok := v.(map[string]interface{}); ok {
			pkg.Config = vStr
		} else {
			return nil, fmt.Errorf("Invalid `config` field type")
		}
	} else {
		pkg.Config = make(map[string]interface{})
	}

	return pkg, nil
}
