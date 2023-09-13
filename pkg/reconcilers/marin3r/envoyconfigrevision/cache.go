package reconcilers

import (
	"context"
	"fmt"
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	envoy_serializer "github.com/3scale-ops/marin3r/pkg/apishelper/serializer"

	marin3rv1alpha1 "github.com/3scale-ops/marin3r/apis/marin3r/v1alpha1"
	xdss "github.com/3scale-ops/marin3r/pkg/discoveryservice/xdss"
	envoy_resources "github.com/3scale-ops/marin3r/pkg/envoy/resources"
	"github.com/3scale-ops/marin3r/pkg/reconcilers/marin3r/envoyconfigrevision/discover"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretCertificate = "tls.crt"
	secretPrivateKey  = "tls.key"
)

type CacheReconciler struct {
	ctx       context.Context
	logger    logr.Logger
	client    client.Client
	xdsCache  xdss.Cache
	decoder   envoy_serializer.ResourceUnmarshaller
	generator envoy_resources.Generator
}

func NewCacheReconciler(ctx context.Context, logger logr.Logger, client client.Client, xdsCache xdss.Cache,
	decoder envoy_serializer.ResourceUnmarshaller, generator envoy_resources.Generator) CacheReconciler {

	return CacheReconciler{ctx, logger, client, xdsCache, decoder, generator}
}

func (r *CacheReconciler) Reconcile(ctx context.Context, req types.NamespacedName, resources []marin3rv1alpha1.Resource,
	nodeID, version string) (*marin3rv1alpha1.VersionTracker, error) {

	snap, err := r.GenerateSnapshot(req, resources)

	if err != nil {
		return nil, err
	}

	oldSnap, err := r.xdsCache.GetSnapshot(nodeID)
	if err != nil || areDifferent(snap, oldSnap) {

		r.logger.Info("Writing new snapshot to xDS cache", "Revision", version, "NodeID", nodeID)
		if err := r.xdsCache.SetSnapshot(ctx, nodeID, snap); err != nil {
			return nil, err
		}

	}

	return &marin3rv1alpha1.VersionTracker{
		Endpoints:        snap.GetVersion(apishelper.Endpoint),
		Clusters:         snap.GetVersion(apishelper.Cluster),
		Routes:           snap.GetVersion(apishelper.Route),
		ScopedRoutes:     snap.GetVersion(apishelper.ScopedRoute),
		Listeners:        snap.GetVersion(apishelper.Listener),
		Secrets:          snap.GetVersion(apishelper.Secret),
		Runtimes:         snap.GetVersion(apishelper.Runtime),
		ExtensionConfigs: snap.GetVersion(apishelper.ExtensionConfig),
	}, nil
}

func (r *CacheReconciler) GenerateSnapshot(req types.NamespacedName, resources []marin3rv1alpha1.Resource) (xdss.Snapshot, error) {
	snap := r.xdsCache.NewSnapshot()

	endpoints := make([]apishelper.Resource, 0, len(resources))
	clusters := make([]apishelper.Resource, 0, len(resources))
	routes := make([]apishelper.Resource, 0, len(resources))
	scopedRoutes := make([]apishelper.Resource, 0, len(resources))
	listeners := make([]apishelper.Resource, 0, len(resources))
	runtimes := make([]apishelper.Resource, 0, len(resources))
	extensionConfigs := make([]apishelper.Resource, 0, len(resources))
	secrets := make([]apishelper.Resource, 0, len(resources))

	for idx, resourceDefinition := range resources {
		switch resourceDefinition.Type {

		case apishelper.Endpoint:

			if resourceDefinition.GenerateFromEndpointSlices != nil {
				// Endpoint discovery enabled
				endpoint, err := discover.Endpoints(r.ctx, r.client, req.Namespace,
					resourceDefinition.GenerateFromEndpointSlices.ClusterName,
					resourceDefinition.GenerateFromEndpointSlices.TargetPort,
					resourceDefinition.GenerateFromEndpointSlices.Selector,
					r.generator, r.logger)
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, endpoint)

			} else {
				// Raw value provided
				res := r.generator.New(apishelper.Endpoint)
				if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
					return nil,
						resourceLoaderError(
							req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
							fmt.Sprintf("Invalid envoy resource value: '%s'", err),
						)
				}
				endpoints = append(endpoints, res)
			}

		case apishelper.Cluster:
			res := r.generator.New(apishelper.Cluster)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}
			clusters = append(clusters, res)

		case apishelper.Route:
			res := r.generator.New(apishelper.Route)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}
			routes = append(routes, res)

		case apishelper.ScopedRoute:
			res := r.generator.New(apishelper.ScopedRoute)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}
			scopedRoutes = append(scopedRoutes, res)

		case apishelper.Listener:
			res := r.generator.New(apishelper.Listener)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}
			listeners = append(listeners, res)

		case apishelper.Secret:
			s := &corev1.Secret{}
			// The webhook will ensure this pointer is set
			name := *resourceDefinition.GenerateFromTlsSecret
			key := types.NamespacedName{Name: name, Namespace: req.Namespace}
			if err := r.client.Get(r.ctx, key, s); err != nil {
				return nil, fmt.Errorf("%s", err.Error())
			}

			// Validate secret holds a certificate
			if s.Type == "kubernetes.io/tls" {
				var res apishelper.Resource

				switch resourceDefinition.GetBlueprint() {
				case marin3rv1alpha1.TlsCertificate:
					res = r.generator.NewTlsCertificateSecret(name, string(s.Data[secretPrivateKey]), string(s.Data[secretCertificate]))
				case marin3rv1alpha1.TlsValidationContext:
					res = r.generator.NewValidationContextSecret(name, string(s.Data[secretCertificate]))
				}

				secrets = append(secrets, res)

			} else {
				err := resourceLoaderError(
					req, name, field.NewPath("spec", "resources").Index(idx).Child("ref"),
					"Only 'kubernetes.io/tls' type secrets allowed",
				)
				return nil, fmt.Errorf("%s", err.Error())

			}

		case apishelper.Runtime:
			res := r.generator.New(apishelper.Runtime)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}
			runtimes = append(runtimes, res)

		case apishelper.ExtensionConfig:
			res := r.generator.New(apishelper.ExtensionConfig)
			if err := r.decoder.Unmarshal(string(resourceDefinition.Value.Raw), res); err != nil {
				return nil,
					resourceLoaderError(
						req, string(resourceDefinition.Value.Raw), field.NewPath("spec", "resources").Index(idx).Child("value"),
						fmt.Sprintf("Invalid envoy resource value: '%s'", err),
					)
			}
			extensionConfigs = append(extensionConfigs, res)

		default:

		}

	}

	snap.SetResources(apishelper.Endpoint, endpoints)
	snap.SetResources(apishelper.Cluster, clusters)
	snap.SetResources(apishelper.Route, routes)
	snap.SetResources(apishelper.ScopedRoute, scopedRoutes)
	snap.SetResources(apishelper.Listener, listeners)
	snap.SetResources(apishelper.Secret, secrets)
	snap.SetResources(apishelper.Runtime, runtimes)
	snap.SetResources(apishelper.ExtensionConfig, extensionConfigs)

	return snap, nil
}

func resourceLoaderError(req types.NamespacedName, value interface{}, resPath *field.Path, msg string) error {
	return errors.NewInvalid(
		schema.GroupKind{Group: "envoy", Kind: "EnvoyConfig"},
		fmt.Sprintf("%s/%s", req.Namespace, req.Name),
		field.ErrorList{field.Invalid(resPath, value, fmt.Sprint(msg))},
	)
}

func areDifferent(a, b xdss.Snapshot) bool {
	for _, rType := range []apishelper.Type{apishelper.Endpoint, apishelper.Cluster, apishelper.Route, apishelper.ScopedRoute,
		apishelper.Listener, apishelper.Secret, apishelper.Runtime, apishelper.ExtensionConfig} {
		if a.GetVersion(rType) != b.GetVersion(rType) {
			return true
		}
	}
	return false
}
