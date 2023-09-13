package envoy

import (
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	envoy_resources_v3 "github.com/3scale-ops/marin3r/pkg/envoy/resources/v3"
)

// Generator in an interface with methods to generate
// envoy resource structs
type Generator interface {
	New(rType apishelper.Type) apishelper.Resource
	NewTlsCertificateSecret(string, string, string) apishelper.Resource
	NewValidationContextSecret(string, string) apishelper.Resource
	NewSecretFromPath(string, string, string) apishelper.Resource
	NewClusterLoadAssignment(string, ...apishelper.UpstreamHost) apishelper.Resource
}

// NewGenerator returns a generator struct for the given API version
func NewGenerator(version apishelper.APIVersion) Generator {

	return envoy_resources_v3.Generator{}
}
