package generators

import (
	"fmt"

	marin3rv1alpha1 "github.com/3scale-ops/marin3r/apis/marin3r/v1alpha1"
	defaults "github.com/3scale-ops/marin3r/pkg/envoy/bootstrap/defaults"
	"github.com/3scale-ops/marin3r/pkg/reconcilers/lockedresources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cfg *GeneratorOptions) EnvoyBootstrap() lockedresources.GeneratorFunction {

	return func() client.Object {

		return &marin3rv1alpha1.EnvoyBootstrap{
			TypeMeta: metav1.TypeMeta{
				Kind:       marin3rv1alpha1.EnvoyBootstrapKind,
				APIVersion: marin3rv1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.resourceName(),
				Namespace: cfg.Namespace,
				Labels:    cfg.labels(),
			},
			Spec: marin3rv1alpha1.EnvoyBootstrapSpec{
				DiscoveryService: cfg.InstanceName,
				ClientCertificate: marin3rv1alpha1.ClientCertificate{
					Directory:  defaults.EnvoyTLSBasePath,
					SecretName: defaults.SidecarClientCertificate,
					Duration: metav1.Duration{
						Duration: cfg.ClientCertificateDuration,
					},
				},
				EnvoyStaticConfig: marin3rv1alpha1.EnvoyStaticConfig{
					ConfigMapNameV2:       defaults.SidecarBootstrapConfigMapV2,
					ConfigMapNameV3:       defaults.SidecarBootstrapConfigMapV3,
					ConfigFile:            fmt.Sprintf("%s/%s", defaults.EnvoyConfigBasePath, defaults.EnvoyConfigFileName),
					ResourcesDir:          defaults.EnvoyConfigBasePath,
					RtdsLayerResourceName: "runtime",
					AdminBindAddress:      "0.0.0.0:9901",
					AdminAccessLogPath:    "/dev/null",
				},
			},
		}
	}
}