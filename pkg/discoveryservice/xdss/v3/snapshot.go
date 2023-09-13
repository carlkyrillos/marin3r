package discoveryservice

import (
	reconcilerutil "github.com/3scale-ops/basereconciler/util"
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	envoy_serializer "github.com/3scale-ops/marin3r/pkg/apishelper/serializer"
	xdss "github.com/3scale-ops/marin3r/pkg/discoveryservice/xdss"
	envoy_resources_v3 "github.com/3scale-ops/marin3r/pkg/envoy/resources/v3"
	cache_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// Snapshot implements "github.com/3scale-ops/marin3r/pkg/discoveryservice/xdss".Snapshot for envoy API v3.
type Snapshot struct {
	v3 *cache_v3.Snapshot
}

// NewSnapshot returns a Snapshot object
func NewSnapshot() Snapshot {

	snap, _ := cache_v3.NewSnapshot("",
		map[resource_v3.Type][]cache_types.Resource{
			resource_v3.EndpointType:        {},
			resource_v3.ClusterType:         {},
			resource_v3.RouteType:           {},
			resource_v3.ScopedRouteType:     {},
			resource_v3.VirtualHostType:     {},
			resource_v3.ListenerType:        {},
			resource_v3.SecretType:          {},
			resource_v3.RuntimeType:         {},
			resource_v3.ExtensionConfigType: {},
		},
	)

	return Snapshot{v3: snap}
}

// Consistent check verifies that the dependent resources are exactly listed in the
// snapshot:
// - all EDS resources are listed by name in CDS resources
// - all RDS resources are listed by name in LDS resources
//
// Note that clusters and listeners are requested without name references, so
// Envoy will accept the snapshot list of clusters as-is even if it does not match
// all references found in xDS.
func (s Snapshot) Consistent() error {
	return s.v3.Consistent()
}

func (s Snapshot) SetResources(rType apishelper.Type, resources []apishelper.Resource) xdss.Snapshot {

	items := make([]cache_types.Resource, 0, len(resources))
	for _, r := range resources {
		items = append(items, cache_types.Resource(r))
	}

	cv3resources := cache_v3.NewResources("", items)
	s.v3.Resources[v3CacheResources(rType)] = cv3resources

	s.SetVersion(rType, s.recalculateVersion(rType))

	return s
}

// GetResources selects snapshot resources by type.
func (s Snapshot) GetResources(rType apishelper.Type) map[string]apishelper.Resource {

	typeURLs := envoy_resources_v3.Mappings()
	resources := map[string]apishelper.Resource{}
	for k, v := range s.v3.GetResources(typeURLs[rType]) {
		resources[k] = v.(apishelper.Resource)
	}
	return resources
}

// GetVersion returns the version for a resource type.
func (s Snapshot) GetVersion(rType apishelper.Type) string {
	typeURLs := envoy_resources_v3.Mappings()
	return s.v3.GetVersion(typeURLs[rType])
}

// SetVersion sets the version for a resource type.
func (s Snapshot) SetVersion(rType apishelper.Type, version string) {
	s.v3.Resources[v3CacheResources(rType)].Version = version
}

func (s Snapshot) recalculateVersion(rType apishelper.Type) string {
	resources := map[string]string{}
	encoder := envoy_serializer.NewResourceMarshaller(envoy_serializer.JSON, apishelper.APIv3)
	for n, r := range s.v3.Resources[v3CacheResources(rType)].Items {
		j, _ := encoder.Marshal(r.Resource)
		resources[n] = string(j)
	}
	if len(resources) > 0 {
		return reconcilerutil.Hash(resources)
	}
	return ""
}

func v3CacheResources(rType apishelper.Type) int {
	types := map[apishelper.Type]int{
		apishelper.Endpoint:        int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.Endpoint])),
		apishelper.Cluster:         int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.Cluster])),
		apishelper.Route:           int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.Route])),
		apishelper.ScopedRoute:     int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.ScopedRoute])),
		apishelper.VirtualHost:     int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.VirtualHost])),
		apishelper.Listener:        int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.Listener])),
		apishelper.Secret:          int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.Secret])),
		apishelper.Runtime:         int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.Runtime])),
		apishelper.ExtensionConfig: int(cache_v3.GetResponseType(envoy_resources_v3.Mappings()[apishelper.ExtensionConfig])),
	}

	return types[rType]
}
