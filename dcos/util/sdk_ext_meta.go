/**
 * The Metadata Extension to the SDK API provides a user-friendly interface to
 * the DC/OS ZooKeeper instance. More specifically it allows you to store and
 * retrieve arbitrary key/value configuration parameters.
 *
 * The data is stored on the same ZK tree as the SDK service and it's deleted
 * when the SDK service is removed.
 */

package util

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/**
 * GetAllMeta returns a map with all the stored configuration properties for this SDK app
 */
func (client *SDKApiClient) GetAllMeta() (map[string]interface{}, error) {

	// Prepare the exchibitor URL to query for fetching the node data
	path := fmt.Sprintf("/dcos-service-%s/Properties", strings.Replace(client.AppID, "/", "__", -1))
	url := fmt.Sprintf(
		"%s/exhibitor/exhibitor/v1/explorer/node-data?key=%s&_=%d",
		client.ClusterURL,
		url.QueryEscape(path),
		int32(time.Now().Unix()),
	)

	log.Printf("[TRACE] Placing GET request to %s", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare request: %s", err.Error())
	}
	for key, value := range client.Headers {
		request.Header.Add(key, value)
	}

	response, err := client.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Unable to place request: %s", err.Error())
	}
	defer response.Body.Close()

	// Exhibitor responds with a hex-encoded byte response. Convert it to bytes
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body: %s", err.Error())
	}

	// Read JSON response
	log.Printf("[TRACE] Received: %s", string(body))
	jsonResponse := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse response as JSON: %s", err.Error())
	}
	log.Printf("[TRACE] Parsed JSON response %v", jsonResponse)

	if bytesBody, ok := jsonResponse["bytes"]; ok {
		configBytes, err := hex.DecodeString(strings.Replace(bytesBody.(string), " ", "", -1))
		if err != nil {
			return nil, fmt.Errorf("Unable to parse response body: %s", err.Error())
		}

		log.Printf("[TRACE] Parsed hex contents to: %s", string(configBytes))

		// Parse configuration body as JSON
		resp := make(map[string]interface{})
		if len(configBytes) == 0 {
			log.Printf("[TRACE] Config is blank")
			return resp, nil
		}
		err = json.Unmarshal(configBytes, &resp)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse the configuration content: %s", err.Error())
		}

		log.Printf("[TRACE] Unmarshalled to %v", resp)
		return resp, nil
	}

	return nil, fmt.Errorf("Missing field `bytes` on response from exhibitor")
}

/**
 * SetAllMeta replaces the entire configuration properties of the SDK app
 */
func (client *SDKApiClient) SetAllMeta(meta map[string]interface{}) error {

	// Prepare the exchibitor URL to query for fetching the node data
	url := fmt.Sprintf(
		"%s/exhibitor/exhibitor/v1/explorer/znode/dcos-service-%s/Properties",
		client.ClusterURL,
		strings.Replace(client.AppID, "/", "__", -1),
	)

	// Serialize configuration to JSON
	log.Printf("[TRACE] Marshalling %v", meta)
	configBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("Unable to marshal configuration data: %s", err.Error())
	}

	payload := fmt.Sprintf("% x", configBytes)
	log.Printf("[TRACE] Placing PUT request to %s with data: %s", url, payload)
	request, err := http.NewRequest("PUT", url, bytes.NewReader([]byte(payload)))
	if err != nil {
		return fmt.Errorf("Unable to prepare request: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	for key, value := range client.Headers {
		request.Header.Add(key, value)
	}

	response, err := client.Client.Do(request)
	if err != nil {
		return fmt.Errorf("Unable to place request: %s", err.Error())
	}
	defer response.Body.Close()

	// Exhibitor responds with a hex-encoded byte response. Convert it to bytes
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Unable to read response body: %s", err.Error())
	}

	// Read JSON response
	log.Printf("[TRACE] Received: %s", string(body))
	jsonResponse := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return fmt.Errorf("Unable to parse response as JSON: %s", err.Error())
	}
	log.Printf("[TRACE] Parsed JSON response %v", jsonResponse)

	// Check for success
	succeeded, ok := jsonResponse["succeeded"]
	if !ok {
		return fmt.Errorf("Unexpected response: Missing `succeeded`")
	}
	if b, ok := succeeded.(bool); ok {
		if !b {
			message, ok := jsonResponse["message"]
			if !ok {
				return fmt.Errorf("Operation did not succeed")
			} else {
				return fmt.Errorf("Operation failed: %s", message)
			}
		}
	} else {
		return fmt.Errorf("Unexpected response: Invalid `succeeded` type")
	}

	return nil
}

/**
 * GetMeta returns a single meta-data parameter value
 */
func (client *SDKApiClient) GetMeta(key string, defaultValue interface{}) (interface{}, error) {
	log.Printf("[TRACE] Getting meta '%s' with default '%v'", key, defaultValue)

	dict, err := client.GetAllMeta()
	if err != nil {
		return nil, fmt.Errorf("Could not get property '%s': %s", key, err.Error())
	}

	if v, ok := dict[key]; ok {
		log.Printf("[TRACE] Key found: %v", v)
		return v, nil
	}

	log.Printf("[TRACE] Key not found in dict, returning default")
	return defaultValue, nil
}

/**
 * SetMeta updates a single meta-data parameter value
 */
func (client *SDKApiClient) SetMeta(key string, value interface{}) error {
	log.Printf("[TRACE] Setting meta '%s' to '%v'", key, value)

	dict, err := client.GetAllMeta()
	if err != nil {
		log.Printf("[WARN] Failed to GetAll: %s", err.Error())
		return fmt.Errorf("Could not fetch state while updating property '%s': %s", key, err.Error())
	}

	dict[key] = value

	err = client.SetAllMeta(dict)
	if err != nil {
		log.Printf("[WARN] Failed to SetAll: %s", err.Error())
		return fmt.Errorf("Could not update state while updating property '%s': %s", key, err.Error())
	}

	return nil
}
