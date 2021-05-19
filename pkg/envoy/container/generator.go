package container

import (
	"fmt"

	operatorv1alpha1 "github.com/3scale-ops/marin3r/apis/operator.marin3r/v1alpha1"
	"github.com/3scale-ops/marin3r/pkg/envoy/container/defaults"
	"github.com/3scale-ops/marin3r/pkg/envoy/container/shutdownmanager"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ContainerConfig struct {
	Name                   string
	Image                  string
	BootstrapConfigMap     string
	ConfigBasePath         string
	ConfigFileName         string
	ConfigVolume           string
	TLSBasePath            string
	TLSVolume              string
	NodeID                 string
	ClusterID              string
	ClientCertSecret       string
	ExtraArgs              []string
	Resources              corev1.ResourceRequirements
	AdminPort              int32
	Ports                  []corev1.ContainerPort
	LivenessProbe          operatorv1alpha1.ProbeSpec
	ReadinessProbe         operatorv1alpha1.ProbeSpec
	ShutdownManagerEnabled bool
	ShutdownManagerPort    int32
	ShutdownManagerImage   string
}

func (cc *ContainerConfig) Containers() []corev1.Container {

	containers := []corev1.Container{{
		Name:    cc.Name,
		Image:   cc.Image,
		Command: []string{"envoy"},
		Args: func() []string {
			args := []string{"-c",
				fmt.Sprintf("%s/%s", cc.ConfigBasePath, cc.ConfigFileName),
				"--service-node",
				cc.NodeID,
				"--service-cluster",
				cc.ClusterID,
			}
			if len(cc.ExtraArgs) > 0 {
				args = append(args, cc.ExtraArgs...)
			}
			return args
		}(),
		Resources: cc.Resources,
		Ports: append(cc.Ports, corev1.ContainerPort{
			Name:          "admin",
			ContainerPort: cc.AdminPort,
		}),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      cc.TLSVolume,
				ReadOnly:  true,
				MountPath: cc.TLSBasePath,
			},
			{
				Name:      cc.ConfigVolume,
				ReadOnly:  true,
				MountPath: cc.ConfigBasePath,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.IntOrString{IntVal: cc.AdminPort},
				},
			},
			InitialDelaySeconds: cc.LivenessProbe.InitialDelaySeconds,
			TimeoutSeconds:      cc.LivenessProbe.TimeoutSeconds,
			PeriodSeconds:       cc.LivenessProbe.PeriodSeconds,
			SuccessThreshold:    cc.LivenessProbe.SuccessThreshold,
			FailureThreshold:    cc.LivenessProbe.FailureThreshold,
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.IntOrString{IntVal: cc.AdminPort},
				},
			},
			InitialDelaySeconds: cc.ReadinessProbe.InitialDelaySeconds,
			TimeoutSeconds:      cc.ReadinessProbe.TimeoutSeconds,
			PeriodSeconds:       cc.ReadinessProbe.PeriodSeconds,
			SuccessThreshold:    cc.ReadinessProbe.SuccessThreshold,
			FailureThreshold:    cc.ReadinessProbe.FailureThreshold,
		},
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		ImagePullPolicy:          corev1.PullIfNotPresent,
	}}

	if cc.ShutdownManagerEnabled {
		containers = append(containers, corev1.Container{
			Name:  "envoy-shtdn-mgr",
			Image: cc.ShutdownManagerImage,
			Args: []string{
				"shutdown-manager",
				"--port",
				fmt.Sprintf("%d", cc.ShutdownManagerPort),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(defaults.ShtdnMgrDefaultCPURequests),
					corev1.ResourceMemory: resource.MustParse(defaults.ShtdnMgrDefaultMemoryRequests),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(defaults.ShtdnMgrDefaultCPULimits),
					corev1.ResourceMemory: resource.MustParse(defaults.ShtdnMgrDefaultMemoryLimits),
				},
			},
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   shutdownmanager.HealthEndpoint,
						Port:   intstr.FromInt(int(cc.ShutdownManagerPort)),
						Scheme: corev1.URISchemeHTTP,
					},
				},
				InitialDelaySeconds: 3,
				PeriodSeconds:       10,
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   shutdownmanager.DrainEndpoint,
						Port:   intstr.FromInt(int(cc.ShutdownManagerPort)),
						Scheme: corev1.URISchemeHTTP,
					},
				},
			},
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			ImagePullPolicy:          corev1.PullIfNotPresent,
		})

		containers[0].Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   shutdownmanager.ShutdownEndpoint,
					Port:   intstr.FromInt(int(cc.ShutdownManagerPort)),
					Scheme: corev1.URISchemeHTTP,
				},
			},
		}
	}

	return containers
}

func (cc *ContainerConfig) Volumes() []corev1.Volume {

	return []corev1.Volume{
		{
			Name: cc.TLSVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: cc.ClientCertSecret,
				},
			},
		},
		{
			Name: cc.ConfigVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cc.BootstrapConfigMap,
					},
				},
			},
		},
	}
}