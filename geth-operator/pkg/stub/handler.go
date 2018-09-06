package stub

import (
	"context"
	"fmt"
	"reflect"

	v1alpha1 "github.com/operator-framework/operator-sdk-samples/geth-operator/pkg/apis/geth/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.GethNode:
		geth := o

		// Ignore the delete event since the garbage collector will clean up all
		// secondary resources for the CR. All secondary resources must have
		// the CR set as their OwnerReference for this to be the case
		if event.Deleted {
			return nil
		}

		// Create the deployment if it doesn't exist
		dep := deploymentForGethNode(geth)
		err := sdk.Create(dep)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment: %v", err)
		}

		// Create the geth client service
		service := serviceForGethNode(geth)
		err = sdk.Create(service)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service: %v", err)
		}

		// Ensure the deployment size is the same as the spec
		err = sdk.Get(dep)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %v", err)
		}
		size := geth.Spec.Size
		if *dep.Spec.Replicas != size {
			dep.Spec.Replicas = &size
			err = sdk.Update(dep)
			if err != nil {
				return fmt.Errorf("failed to update deployment: %v", err)
			}
		}

		// Update the GethNode status with the pod names
		podList := podList()
		ls := labelsForGethNode(geth)
		labelSelector := labels.SelectorFromSet(ls).String()
		listOps := &metav1.ListOptions{LabelSelector: labelSelector}
		err = sdk.List(geth.Namespace, podList, sdk.WithListOptions(listOps))
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}
		podNames := getPodNames(podList.Items)
		if !reflect.DeepEqual(podNames, geth.Status.Nodes) {
			geth.Status.Nodes = podNames
			err = sdk.Update(geth)
			if err != nil {
				return fmt.Errorf("failed to update geth status: %v", err)
			}
		}
	}
	return nil
}

func newGethNodeConfigMap(cr *v1alpha1.GethNode) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"config.toml": `|
			[Eth]
			NetworkId = 1
			SyncMode = "fast"
			LightPeers = 20
			DatabaseCache = 128
			GasPrice = 18000000000
			EthashCacheDir = "ethash"
			EthashCachesInMem = 2
			EthashCachesOnDisk = 3
			EthashDatasetDir = "/etc/geth/.ethereum/ethash"
			EthashDatasetsInMem = 1
			EthashDatasetsOnDisk = 2
			EnablePreimageRecording = false

			[Node]
			DataDir = "/eth/geth/.ethereum"

			[Node.P2P]
			MaxPeers = 25
			NoDiscovery = false
			DiscoveryV5Addr = ":30304"
			ListenAddr = ":30303"
			EnableMsgEvents = false
			`,
		},
	}
}

// labelsForGethNode returns the labels for selecting the resources
// belonging to the given geth CR name.
func labelsForGethNode(cr *v1alpha1.GethNode) map[string]string {
	return map[string]string{
		"app":     "geth",
		"geth_cr": cr.Name,
	}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// asOwner returns an OwnerReference set as the geth CR
func asOwner(m *v1alpha1.GethNode) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: m.APIVersion,
		Kind:       m.Kind,
		Name:       m.Name,
		UID:        m.UID,
		Controller: &trueVar,
	}
}

func serviceForGethNode(cr *v1alpha1.GethNode) *corev1.Service {
	ls := labelsForGethNode(cr)

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Name:       "listen",
					Protocol:   corev1.ProtocolUDP,
					Port:       30303,
					TargetPort: intstr.FromString("p2p-listen"),
				},
				{
					Name:       "discovery",
					Port:       30304,
					TargetPort: intstr.FromString("p2p-discovery"),
				},
				{
					Name:       "rpc-json",
					Port:       8545,
					TargetPort: intstr.FromString("rpc-json"),
				},
				{
					Name:       "rpc-ws",
					Port:       8546,
					TargetPort: intstr.FromString("rpc-ws"),
				},
			},
		},
	}
}

func deploymentForGethNode(cr *v1alpha1.GethNode) *appsv1.Deployment {
	ls := labelsForGethNode(cr)
	replicas := cr.Spec.Size

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "geth",
							Image: "ethereum/client-go:alpine",
							Args:  []string{"--config", "/etc/geth/cfg/config.toml", "--datadir", "/etc/geth/.ethereum"},
							Ports: []corev1.ContainerPort{
								{
									Name:          "p2p-listen",
									ContainerPort: 30303,
									Protocol:      corev1.ProtocolUDP,
								},
								{
									Name:          "p2p-discovery",
									ContainerPort: 30304,
								},
								{
									Name:          "rpc-json",
									ContainerPort: 8545,
								},
								{
									Name:          "rpc-ws",
									ContainerPort: 8546,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/etc/geth/.ethereum",
								},
								{
									Name:      "config",
									MountPath: "/etc/geth/cfg",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "datadir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "$(HOME)/.ethereum",
								},
							},
						},
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "geth",
									},
									Items: []corev1.KeyToPath{{
										Key:  "config.toml",
										Path: "config.toml",
									}},
								},
							},
						},
					},
				},
			},
		},
	}

	addOwnerRefToObject(dep, asOwner(cr))
	return dep
}

// podList returns a v1.PodList object
func podList() *corev1.PodList {
	return &corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
