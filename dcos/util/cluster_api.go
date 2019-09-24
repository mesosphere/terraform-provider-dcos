package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dcos/client-go/dcos"
)

type DCOSVersionSpec struct {
	Version         string `json:"version,omitempty"`
	DcosVariant     string `json:"dcos-variant,omitempty"`
	DcosImageCommit string `json:"dcos-image-commit,omitempty"`
	BootstrapId     string `json:"bootstrap-id,omitempty"`
}

func DCOSHTTPClient(client *dcos.APIClient) *http.Client {
	return client.HTTPClient()
}

func DCOSNewRequest(client *dcos.APIClient, method, url string, body io.Reader) (*http.Request, error) {
	config := client.CurrentDCOSConfig()
	request, err := http.NewRequest(method, fmt.Sprintf("%s%s", config.URL(), url), body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("token=%s", config.ACSToken()))
	return request, nil
}

/**
 * Get the DC/OS version from /dcos-metadata/dcos-version.json
 */
func DCOSGetVersion(client *dcos.APIClient) (DCOSVersionSpec, error) {
	var ver DCOSVersionSpec

	http := DCOSHTTPClient(client)
	req, err := DCOSNewRequest(client, "GET", "/dcos-metadata/dcos-version.json", nil)
	if err != nil {
		return ver, fmt.Errorf("Unable to create request: %s", err.Error())
	}

	resp, err := http.Do(req)
	if err != nil {
		return ver, fmt.Errorf("Unable to place request: %s", err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ver, fmt.Errorf("Unable to read response: %s", err.Error())
	}

	err = json.Unmarshal(body, &ver)
	if err != nil {
		return ver, fmt.Errorf("Unable to parse response: %s", err.Error())
	}

	return ver, nil
}
