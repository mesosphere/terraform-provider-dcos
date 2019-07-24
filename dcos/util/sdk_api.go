package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dcos/client-go/dcos"
)

type SDKApiClient struct {
	AppID   string
	BaseURL string
	Client  *http.Client
	Headers map[string]string
}

/**
 * CreateSDKAPIClientFor initializes an SDKApiClient API
 */
func CreateSDKAPIClient(client *dcos.APIClient, appId string) *SDKApiClient {
	config := client.CurrentDCOSConfig()
	headers := map[string]string{
		"Authorization": fmt.Sprintf("token=%s", config.ACSToken()),
	}

	return &SDKApiClient{
		AppID:   appId,
		BaseURL: fmt.Sprintf("%s/service/%s", config.URL(), appId),
		Client:  client.HTTPClient(),
		Headers: headers,
	}
}

func (client *SDKApiClient) unmarshalJsonResponse(response *http.Response, respBody interface{}) (*http.Response, error) {
	// Create a streaming JSON decoder
	jsonReader := json.NewDecoder(response.Body)
	err := jsonReader.Decode(&respBody)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse response as JSON: %s", err.Error())
	}

	return response, nil
}

/**
 * requestPOST places a POST request to the service endpoint
 */
func (client *SDKApiClient) postJSON(endpoint string, reqBody interface{}, respBody interface{}) (*http.Response, error) {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Unable to serialize body: %s", err.Error())
	}

	url := fmt.Sprintf("%s/%s", client.BaseURL, endpoint)
	log.Printf("[TRACE] Placing POST request to %s with data: %s", url, string(payload))
	request, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare request: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	for key, value := range client.Headers {
		request.Header.Add(key, value)
	}

	response, err := client.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Unable to place request: %s", err.Error())
	}
	defer response.Body.Close()

	log.Printf("[TRACE] Server responded with %s", response.Status)
	return client.unmarshalJsonResponse(response, respBody)
}

/**
 * requestGET places a GET request to the service endpoint
 */
func (client *SDKApiClient) getJSON(endpoint string, respBody interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", client.BaseURL, endpoint)
	log.Printf("[TRACE] Placing GET request to %s", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare request: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	for key, value := range client.Headers {
		request.Header.Add(key, value)
	}

	response, err := client.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Unable to place request: %s", err.Error())
	}
	defer response.Body.Close()

	log.Printf("[TRACE] Server responded with %s", response.Status)
	return client.unmarshalJsonResponse(response, respBody)
}
