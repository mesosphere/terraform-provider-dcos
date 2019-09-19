package dcos

import (
	"github.com/dcos/client-go/dcos"
	"github.com/mesosphere/terraform-provider-dcos/dcos/util"
)

/**
 * The provide state, as created during instantiation.
 * Since this object will remain allocated until the provider is destroyed,
 * you can use for keeping cached information, for speeding-up the process.
 */
type ProviderState struct {
	Client     *dcos.APIClient
	CliWrapper *util.CliWrapper
}
