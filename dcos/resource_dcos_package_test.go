package dcos

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDcosPackage_import(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ResourceName:      "dcos_package.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
