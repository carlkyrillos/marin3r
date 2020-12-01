package controllers

import (
	"context"
	"fmt"
	"time"

	marin3rv1alpha1 "github.com/3scale/marin3r/apis/marin3r/v1alpha1"
	operatorv1alpha1 "github.com/3scale/marin3r/apis/operator.marin3r/v1alpha1"
	"github.com/3scale/marin3r/pkg/webhooks/podv1mutator"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileEnabledNamespaces is in charge of keep the resources that envoy sidecars require available in all
// the active namespaces:
//     - an EnvoyBootstrap resource
//     - a label to enable sidecar injection in the namespace
func (r *DiscoveryServiceReconciler) reconcileEnabledNamespaces(ctx context.Context, log logr.Logger) (reconcile.Result, error) {
	errList := []error{}
	// Reconcile each namespace in the list of enabled namespaces
	for _, ns := range r.ds.Spec.EnabledNamespaces {
		err := r.reconcileEnabledNamespace(ctx, ns, log)
		if err != nil {
			errList = append(errList, err)
		}
		// Keep going even if an error is returned
	}

	if len(errList) != 0 {
		return reconcile.Result{}, fmt.Errorf("Failed reconciling enabled namespaces: %v", errList)
	}

	return reconcile.Result{}, nil
}

func (r *DiscoveryServiceReconciler) reconcileEnabledNamespace(ctx context.Context, namespace string, log logr.Logger) error {

	ns := &corev1.Namespace{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: namespace}, ns)

	if err != nil {
		// Namespace should exist
		return err
	}

	ok, err := isSidecarEnabled(r.ds, ns)
	if err != nil {
		return err
	}

	if !ok {

		patch := client.MergeFrom(ns.DeepCopy())

		// Init label's map
		if ns.GetLabels() == nil {
			ns.SetLabels(map[string]string{})
		}

		// Set namespace labels
		ns.ObjectMeta.Labels[operatorv1alpha1.DiscoveryServiceEnabledKey] = operatorv1alpha1.DiscoveryServiceEnabledValue
		ns.ObjectMeta.Labels[operatorv1alpha1.DiscoveryServiceLabelKey] = r.ds.GetName()

		if err := r.Client.Patch(ctx, ns, patch); err != nil {
			return err
		}
		log.Info("Patched Namespace", "Namespace", namespace)
	}

	eb := &marin3rv1alpha1.EnvoyBootstrap{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: r.ds.GetName(), Namespace: namespace}, eb); err != nil {

		if errors.IsNotFound(err) {
			eb, err := genEnvoyBootstrapObject(namespace, r.ds)
			if err != nil {
				return err
			}
			if err := controllerutil.SetControllerReference(r.ds, eb, r.Scheme); err != nil {
				return err
			}
			if err := r.Client.Create(ctx, eb); err != nil {
				return err
			}
			log.Info("Created EnvoyBootstrap", "Name", r.ds.GetName(), "Namespace", namespace)
			return nil
		}
		return err
	}

	return nil
}

func isSidecarEnabled(owner metav1.Object, object metav1.Object) (bool, error) {

	value, ok := object.GetLabels()[operatorv1alpha1.DiscoveryServiceLabelKey]
	if ok {
		if value == owner.GetName() {
			return true, nil
		}
		return false, fmt.Errorf("Namespace already onwed by %s", value)
	}

	return false, nil
}

func genEnvoyBootstrapObject(namespace string, ds *operatorv1alpha1.DiscoveryService) (*marin3rv1alpha1.EnvoyBootstrap, error) {

	duration, err := time.ParseDuration("48h")
	if err != nil {
		return nil, err
	}

	return &marin3rv1alpha1.EnvoyBootstrap{
		ObjectMeta: metav1.ObjectMeta{Name: ds.GetName(), Namespace: namespace},
		Spec: marin3rv1alpha1.EnvoyBootstrapSpec{
			DiscoveryService: ds.GetName(),
			ClientCertificate: &marin3rv1alpha1.ClientCertificate{
				Directory:  podv1mutator.DefaultEnvoyTLSBasePath,
				SecretName: podv1mutator.DefaultClientCertificate,
				Duration: metav1.Duration{
					Duration: duration,
				},
			},
			EnvoyStaticConfig: &marin3rv1alpha1.EnvoyStaticConfig{
				ConfigMapNameV2:       podv1mutator.DefaultBootstrapConfigMapV2,
				ConfigMapNameV3:       podv1mutator.DefaultBootstrapConfigMapV3,
				ConfigFile:            fmt.Sprintf("%s/%s", podv1mutator.DefaultEnvoyConfigBasePath, podv1mutator.DefaultEnvoyConfigFileName),
				ResourcesDir:          podv1mutator.DefaultEnvoyConfigBasePath,
				RtdsLayerResourceName: "runtime",
				AdminBindAddress:      "0.0.0.0:9901",
				AdminAccessLogPath:    "/dev/null",
			},
		},
	}, nil
}