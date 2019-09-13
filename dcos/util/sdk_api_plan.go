package util

import (
	"fmt"
)

type PlanStep struct {
	Id      string `json:"id"`
	Status  string `json:"status"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

type PlanPhase struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Steps    []PlanStep `json:"steps"`
	Strategy string     `json:"strategy"`
	Status   string     `json:"status"`
}

type PlansListResponse struct {
	Phases   []PlanPhase `json:"phases"`
	Strategy string      `json:"strategy"`
	Status   string      `json:"status"`
}

type PlanRestartRequest struct {
	Phase string `json:"phase"`
	Step  string `json:"step"`
}

/**
 * Describe package
 */
func (client *SDKApiClient) PlanGetStatus(plan string) (*PlansListResponse, error) {
	var jResp PlansListResponse
	_, err := client.getJSON(fmt.Sprintf("v1/plans/%s", plan), &jResp)
	if err != nil {
		return nil, fmt.Errorf("Unable to place GET request: %s", err.Error())
	}

	return &jResp, nil
}

/**
 * Describe package
 */
func (client *SDKApiClient) PlanRestart(plan string) error {
	var jResp map[string]interface{}
	var jReq PlanRestartRequest

	_, err := client.postJSON(fmt.Sprintf("v1/plans/%s/restart", plan), &jReq, &jResp)
	if err != nil {
		return fmt.Errorf("Unable to place plan restart POST request: %s", err.Error())
	}

	return nil
}
