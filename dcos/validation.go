package dcos

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
)

// validateRegexp is borrowed from https://github.com/terraform-providers/terraform-provider-google/blob/c5bbdce38eb1a971c95691ee3d9f26efca1d595e/google/validation.go#L73-L83
func validateRegexp(re string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)
		if !regexp.MustCompile(re).MatchString(value) {
			errors = append(errors, fmt.Errorf(
				"%q (%q) doesn't match regexp %q", k, value, re))
		}

		return
	}
}
