package stub

import (
	"context"

	"github.com/operator-framework/operator-sdk-samples/dex-operator/pkg/apis/auth/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Dex:
		if event.Deleted {
			return nil
		}
		err := sdk.Create(newDexPod(o))
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create dex pod : %v", err)
			return err
		}
		err = sdk.Create(newDexPod(o))
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create dex pod : %v", err)
			return err
		}
		err = sdk.Create(newDexService(o))
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create dex pod : %v", err)
			return err
		}
	}
	return nil
}

func newDexService(cr *v1alpha1.Dex) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "dex"},
			Ports: []corev1.ServicePort{corev1.ServicePort{
				Name:       "dex",
				Protocol:   corev1.ProtocolTCP,
				Port:       5556,
				TargetPort: 5556,
				NodePort:   32000,
			}},
		},
	}
}

func newDexConfigMap(cr *v1alpha1.Dex) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"config.yaml": `|
			issuer: https://dex.example.com:32000
			storage:
			  type: kubernetes
			  config:
				inCluster: true
			web:
			  https: 0.0.0.0:5556
			  tlsCert: /etc/dex/tls/tls.crt
			  tlsKey: /etc/dex/tls/tls.key
			connectors:
			- type: github
			  id: github
			  name: GitHub
			  config:
				clientID: $GITHUB_CLIENT_ID
				clientSecret: $GITHUB_CLIENT_SECRET
				redirectURI: https://dex.example.com:32000/callback
				org: kubernetes
			oauth2:
			  skipApprovalScreen: true
			staticClients:
			- id: example-app
			  redirectURIs:
			  - 'http://127.0.0.1:5555/callback'
			  name: 'Example App'
			  secret: ZXhhbXBsZS1hcHAtc2VjcmV0
			enablePasswordDB: true
			staticPasswords:
			- email: "admin@example.com"
			  # bcrypt hash of the string "password"
			  hash: "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
			  username: "admin"
			  userID: "08a8684b-db88-4b73-90a9-3cd1661f5466"
			`,
		},
	}
}

// newbusyBoxPod demonstrates how to create a busybox pod
func newDexPod(cr *v1alpha1.Dex) *appsv1.Deployment {
	labels := map[string]string{
		"app":    "dex",
		"dex-cr": cr.Name,
	}
	repSize := cr.Spec.Size
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Dex",
				}),
			},
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &repSize,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "dex",
							Image:   "quay.io/coreos/dex:v2.0.0-beta.1",
							Command: []string{"/usr/local/bin/dex", "serve", "/etc/dex/cfg/config.yaml"},
							Ports: []corev1.ContainerPort{{
								ContainerPort: 5556,
								Name:          "memcached",
							}},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "config",
								MountPath: "/etc/dex/cfg",
							}},
						},
					},
					Volumes: []corev1.Volume{
						{
							corev1.Volume{
								Name: "config",

								corev1.VolumeSource{
									ConfigMap: corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "dex",
										},
										Items: []corev1.KeyToPath{
											corev1.KeyToPath{
												Key:  "config.yaml",
												Path: "config.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
