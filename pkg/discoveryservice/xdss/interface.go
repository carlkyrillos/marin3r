package discoveryservice

import (
	"context"
	"github.com/3scale-ops/marin3r/pkg/apishelper"
)

// Cache is a snapshot-based cache that maintains a single versioned
// snapshot of responses per node. SnapshotCache consistently replies with the
// latest snapshot. For the protocol to work correctly in ADS mode, EDS/RDS
// requests are responded only when all resources in the snapshot xDS response
// are named as part of the request. It is expected that the CDS response names
// all EDS clusters, and the LDS response names all RDS routes in a snapshot,
// to ensure that Envoy makes the request for all EDS clusters or RDS routes
// eventually.
type Cache interface {
	SetSnapshot(context.Context, string, Snapshot) error
	GetSnapshot(string) (Snapshot, error)
	ClearSnapshot(string)
	NewSnapshot() Snapshot
}

// Snapshot is an internally consistent snapshot of xDS resources.
// Consistency is important for the convergence as different resource types
// from the snapshot may be delivered to the proxy in arbitrary order.
type Snapshot interface {
	Consistent() error
	SetResources(apishelper.Type, []apishelper.Resource) Snapshot
	GetResources(apishelper.Type) map[string]apishelper.Resource
	GetVersion(apishelper.Type) string
	SetVersion(apishelper.Type, string)
}
