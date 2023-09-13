package envoy

import (
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	envoy_resources_v3 "github.com/3scale-ops/marin3r/pkg/envoy/resources/v3"
)

func TypeURL(rType apishelper.Type, version apishelper.APIVersion) string {
	return envoy_resources_v3.Mappings()[rType]
}
