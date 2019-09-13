package cosmos

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/**
 * Create a new cosmos repository from a string buffer
 */
func NewRepoFromString(data string) (CosmosRepository, error) {
	rawRepo := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &rawRepo)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse repository JSON: %s", err.Error())
	}

	// Parse repo
	return parseRepo(rawRepo)
}

/**
 * Create a new cosmos repository from the given URL
 */
func NewRepoFromURL(url string) (CosmosRepository, error) {
	client := new(http.Client)

	// Place request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare request: %s", err.Error())
	}
	request.Header.Add("Accept-Encoding", "gzip")

	// Make sure we have the bare minimum headers required by the universe convert
	// script, in order to be able to read convert URLs
	request.Header.Add("Accept", "application/json; version=v3")
	request.Header.Add("User-Agent", "CosmosRepoGo/1.0 (dcos/1.13)")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Unable to place request: %s", err.Error())
	}
	defer response.Body.Close()

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse response gzip stream: %s", err.Error())
		}
		defer reader.Close()
	default:
		reader = response.Body
	}

	// Create a streaming JSON decoder
	jsonReader := json.NewDecoder(reader)
	var rawRepo map[string]interface{}
	err = jsonReader.Decode(&rawRepo)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse response as JSON: %s", err.Error())
	}

	// Parse repo
	return parseRepo(rawRepo)
}
