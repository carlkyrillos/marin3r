package envoy

import (
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// Mappings return a map associating "github.com/3scale-ops/marin3r/pkg/envoy/resources".Type to
// the v3 envoy API type URLs for each resource type
func Mappings() map[apishelper.Type]string {
	return map[apishelper.Type]string{
		apishelper.Listener:        resource_v3.ListenerType,
		apishelper.Route:           resource_v3.RouteType,
		apishelper.ScopedRoute:     resource_v3.ScopedRouteType,
		apishelper.VirtualHost:     resource_v3.VirtualHostType,
		apishelper.Cluster:         resource_v3.ClusterType,
		apishelper.Endpoint:        resource_v3.EndpointType,
		apishelper.Secret:          resource_v3.SecretType,
		apishelper.Runtime:         resource_v3.RuntimeType,
		apishelper.ExtensionConfig: resource_v3.ExtensionConfigType,
	}
}
