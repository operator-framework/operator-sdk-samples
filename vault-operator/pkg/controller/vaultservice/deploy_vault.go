package vaultservice

import (
	"context"
	"fmt"
	"path/filepath"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	vaultConfigVolName   = "vault-config"
	vaultClientPort      = 8200
	vaultClusterPort     = 8201
	vaultClientPortName  = "vault-client"
	vaultClusterPortName = "vault-cluster"
	vaultTLSAssetVolume  = "vault-tls-secret"

	evnVaultRedirectAddr = "VAULT_API_ADDR"
	evnVaultClusterAddr  = "VAULT_CLUSTER_ADDR"

	exporterStatsdPort = 9125
	exporterPromPort   = 9102
	exporterImage      = "prom/statsd-exporter:v0.5.0"
)

// deployVault deploys a vault service.
// deployVault is a multi-steps process. It creates the deployment, the service and
// other related Kubernetes objects for Vault. Any intermediate step can fail.
//
// deployVault is idempotent. If an object already exists, this function will ignore creating
// it and return no error. It is safe to retry on this function.
func (r *ReconcileVaultService) deployVault(v *vaultv1alpha1.VaultService) error {
	selector := LabelsForVault(v.GetName())

	podTempl := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.GetName(),
			Namespace: v.GetNamespace(),
			Labels:    selector,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{vaultContainer(v), statsdExporterContainer()},
			Volumes: []corev1.Volume{{
				Name: vaultConfigVolName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapNameForVault(v),
						},
					},
				},
			}},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    func(i int64) *int64 { return &i }(9000),
				RunAsNonRoot: func(b bool) *bool { return &b }(true),
				FSGroup:      func(i int64) *int64 { return &i }(9000),
			},
		},
	}
	if v.Spec.Pod != nil {
		applyPodPolicy(&podTempl.Spec, v.Spec.Pod)
	}

	configEtcdBackendTLS(&podTempl, v)
	configVaultServerTLS(&podTempl, v)

	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.GetName(),
			Namespace: v.GetNamespace(),
			Labels:    selector,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &v.Spec.Nodes,
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: podTempl,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				},
			},
		},
	}
	addOwnerRefToObject(d, asOwner(v))
	err := r.client.Create(context.TODO(), d)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.GetName(),
			Namespace: v.GetNamespace(),
			Labels:    selector,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:     vaultClientPortName,
					Protocol: corev1.ProtocolTCP,
					Port:     vaultClientPort,
				},
				{
					Name:     vaultClusterPortName,
					Protocol: corev1.ProtocolTCP,
					Port:     vaultClusterPort,
				},
				{
					Name:     "prometheus",
					Protocol: corev1.ProtocolTCP,
					Port:     exporterPromPort,
				},
			},
		},
	}
	addOwnerRefToObject(svc, asOwner(v))
	err = r.client.Create(context.TODO(), svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create vault service: %v", err)
	}
	return nil
}

// configEtcdBackendTLS configures the volume and mounts in vault pod to
// set up etcd backend TLS assets
func configEtcdBackendTLS(pt *corev1.PodTemplateSpec, v *vaultv1alpha1.VaultService) {
	sn := EtcdClientTLSSecretName(v.Name)
	pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
		Name: vaultTLSAssetVolume,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{{
					Secret: &corev1.SecretProjection{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: sn,
						},
					},
				}},
			},
		},
	})
	pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      vaultTLSAssetVolume,
		ReadOnly:  true,
		MountPath: vaultTLSAssetDir,
	})
}

// configVaultServerTLS mounts the volume containing the vault server TLS assets for the vault pod
func configVaultServerTLS(pt *corev1.PodTemplateSpec, v *vaultv1alpha1.VaultService) {
	secretName := v.Spec.TLS.Static.ServerSecret

	serverTLSVolume := corev1.VolumeProjection{
		Secret: &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretName,
			},
		},
	}
	pt.Spec.Volumes[1].VolumeSource.Projected.Sources = append(pt.Spec.Volumes[1].VolumeSource.Projected.Sources, serverTLSVolume)
}

func applyPodPolicy(s *corev1.PodSpec, p *vaultv1alpha1.PodPolicy) {
	for i := range s.Containers {
		s.Containers[i].Resources = p.Resources
	}

	for i := range s.InitContainers {
		s.InitContainers[i].Resources = p.Resources
	}
}

// IsPodReady checks the status of the pod for the Ready condition
func IsPodReady(p corev1.Pod) bool {
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func vaultContainer(v *vaultv1alpha1.VaultService) corev1.Container {
	return corev1.Container{
		Name:  "vault",
		Image: fmt.Sprintf("%s:%s", v.Spec.BaseImage, v.Spec.Version),
		Command: []string{
			"/bin/vault",
			"server",
			"-config=" + vaultConfigPath,
		},
		Env: []corev1.EnvVar{
			{
				Name:  evnVaultRedirectAddr,
				Value: vaultServiceURL(v.GetName(), v.GetNamespace(), vaultClientPort),
			},
			{
				Name:  evnVaultClusterAddr,
				Value: vaultServiceURL(v.GetName(), v.GetNamespace(), vaultClusterPort),
			},
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      vaultConfigVolName,
			MountPath: filepath.Dir(vaultConfigPath),
		}},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				// Vault requires mlock syscall to work.
				// Without this it would fail "Error initializing core: Failed to lock memory: cannot allocate memory"
				Add: []corev1.Capability{"IPC_LOCK"},
			},
		},
		Ports: []corev1.ContainerPort{{
			Name:          vaultClientPortName,
			ContainerPort: int32(vaultClientPort),
		}, {
			Name:          vaultClusterPortName,
			ContainerPort: int32(vaultClusterPort),
		}},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"curl",
						"--connect-timeout", "5",
						"--max-time", "10",
						"-k", "-s",
						fmt.Sprintf("https://localhost:%d/v1/sys/health", vaultClientPort),
					},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       60,
			FailureThreshold:    3,
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/v1/sys/health",
					Port:   intstr.FromInt(vaultClientPort),
					Scheme: corev1.URISchemeHTTPS,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			FailureThreshold:    3,
		},
	}
}

// vaultServiceURL returns the DNS record of the vault service in the given namespace.
func vaultServiceURL(name, namespace string, port int) string {
	return fmt.Sprintf("https://%s.%s.svc:%d", name, namespace, port)
}

func statsdExporterContainer() corev1.Container {
	return corev1.Container{
		Name:  "statsd-exporter",
		Image: exporterImage,
		Ports: []corev1.ContainerPort{{
			Name:          "statsd",
			ContainerPort: exporterStatsdPort,
			Protocol:      corev1.ProtocolUDP,
		}, {
			Name:          "prometheus",
			ContainerPort: exporterPromPort,
			Protocol:      corev1.ProtocolTCP,
		}},
	}
}
