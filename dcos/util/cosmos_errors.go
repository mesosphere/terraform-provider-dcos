package util

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dcos/client-go/dcos"
)

/**
 * GetVerboseCosmosError collects as much information as possible from the error
 * and the HTTP response in order to provide enough information to the user,
 * explaining the error that occurred.
 */
func GetVerboseCosmosError(error error, resp *http.Response) string {
	if apiError, ok := error.(dcos.GenericOpenAPIError); ok {
		if apiError.Model() != nil {
			if cosmosError, ok := apiError.Model().(dcos.CosmosError); ok {
				log.Printf(
					"[TRACE] Formatting CosmosError type='%s', message='%s', data=%s",
					cosmosError.Type, cosmosError.Message, PrintJSON(cosmosError.Data),
				)

				switch cosmosError.Type {
				case "MarathonAppNotFound":
					if appId, ok := cosmosError.Data["appId"]; ok {
						return fmt.Sprintf(
							"A marathon app with name '%s' was not found",
							appId.(string),
						)
					}

				case "PackageAlreadyInstalled":
					return "A package with the same name is already installed"

				case "JsonSchemaMismatch":
					var failures []string = nil
					errors, hasErrors := cosmosError.Data["errors"].([]interface{})
					if !hasErrors {
						break
					}
					for _, err := range errors {
						if errMap, ok := err.(map[string]interface{}); ok {
							if level, ok := errMap["level"]; ok {
								if level.(string) == "error" {
									failures = append(failures, errMap["message"].(string))
								}
							}
						}
					}

					return fmt.Sprintf(
						"%s: %s\n* %s",
						cosmosError.Type,
						cosmosError.Message,
						strings.Join(failures, "\n* "),
					)
				}

				return fmt.Sprintf(
					"%s: %s (data=%s)",
					cosmosError.Type,
					cosmosError.Message,
					PrintJSON(cosmosError.Data),
				)

			}
		}

		return apiError.Error()
	}

	return fmt.Sprintf("Got an HTTP %s response", resp.Status)
}
