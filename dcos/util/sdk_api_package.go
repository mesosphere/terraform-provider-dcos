package util

import (
	"fmt"
)

type updateRequest struct {
	AppID          string                 `json:"appId"`
	PackageVersion string                 `json:"packageVersion,omitempty"`
	OptionsJSON    map[string]interface{} `json:"options,omitempty"`
	Replace        bool                   `json:"replace"`
}

type describeRequest struct {
	AppID string `json:"appId"`
}

type describePackageResponse struct {
	Version string `json:"version"`
}

type describeResponse struct {
	Package      describePackageResponse `json:"package"`
	UpgradesTo   []string                `json:"upgradesTo"`
	DowngradesTo []string                `json:"downgradesTo"`
	// Note: ResolvedOptions is only provided on DC/OS EE 1.10 or later
	ResolvedOptions map[string]interface{} `json:"resolvedOptions"`
}

/**
 * Describe package
 */
func (client *SDKApiClient) PackageDescribe() (*describeResponse, error) {
	var jResp describeResponse
	jReq := describeRequest{client.AppID}

	_, err := client.postJSON("describe", jReq, &jResp)
	if err != nil {
		return nil, fmt.Errorf("Unable to place POST request: %s", err.Error())
	}

	return &jResp, nil
}
