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
		depLabels := labelsForGethNode(geth)
		labelSelector := labels.SelectorFromSet(depLabels).String()
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

			[Eth.TxPool]
			NoLocals = false
			Journal = "transactions.rlp"
			Rejournal = 3600000000000
			PriceLimit = 1
			PriceBump = 10
			AccountSlots = 16
			GlobalSlots = 4096
			AccountQueue = 64
			GlobalQueue = 1024
			Lifetime = 10800000000000

			[Eth.GPO]
			Blocks = 10
			Percentile = 50

			[Shh]
			MaxMessageSize = 1048576
			MinimumAcceptedPOW = 2e-01

			[Node]
			DataDir = "/eth/geth/.ethereum"
			# Uncomment below for HTTP/WebSocket support
			# IPCPath = "geth.ipc"
			# HTTPPort = 8545
			# HTTPModules = ["net", "web3", "eth", "shh"]
			# WSPort = 8546
			# WSModules = ["net", "web3", "eth", "shh"]

			[Node.P2P]
			MaxPeers = 25
			NoDiscovery = false
			DiscoveryV5Addr = ":30304"
			BootstrapNodes = ["enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303", "enode://3f1d12044546b76342d59d4a05532c14b85aa669704bfe1f864fe079415aa2c02d743e03218e57a33fb94523adb54032871a6c51b2cc5514cb7c7e35b3ed0a99@13.93.211.84:30303", "enode://78de8a0916848093c73790ead81d1928bec737d565119932b98c6b100d944b7a95e94f847f689fc723399d2e31129d182f7ef3863f2b4c820abbf3ab2722344d@191.235.84.50:30303", "enode://158f8aab45f6d19c6cbf4a089c2670541a8da11978a2f90dbf6a502a4a3bab80d288afdbeb7ec0ef6d92de563767f3b1ea9e8e334ca711e9f8e2df5a0385e8e6@13.75.154.138:30303", "enode://1118980bf48b0a3640bdba04e0fe78b1add18e1cd99bf22d53daac1fd9972ad650df52176e7c7d89d1114cfef2bc23a2959aa54998a46afcf7d91809f0855082@52.74.57.123:30303", "enode://979b7fa28feeb35a4741660a16076f1943202cb72b6af70d327f053e248bab9ba81760f39d0701ef1d8f89cc1fbd2cacba0710a12cd5314d5e0c9021aa3637f9@5.1.83.226:30303"]
			BootstrapNodesV5 = ["enode://0cc5f5ffb5d9098c8b8c62325f3797f56509bff942704687b6530992ac706e2cb946b90a34f1f19548cd3c7baccbcaea354531e5983c7d1bc0dee16ce4b6440b@40.118.3.223:30305", "enode://1c7a64d76c0334b0418c004af2f67c50e36a3be60b5e4790bdac0439d21603469a85fad36f2473c9a80eb043ae60936df905fa28f1ff614c3e5dc34f15dcd2dc@40.118.3.223:30308", "enode://85c85d7143ae8bb96924f2b54f1b3e70d8c4d367af305325d30a61385a432f247d2c75c45c6b4a60335060d072d7f5b35dd1d4c45f76941f62a4f83b6e75daaf@40.118.3.223:30309"]
			StaticNodes = []
			TrustedNodes = []
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

func deploymentForGethNode(cr *v1alpha1.GethNode) *appsv1.Deployment {
	depLabels := labelsForGethNode(cr)
	replicas := cr.Spec.Size

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    depLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: depLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: depLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "geth",
							Image: "ethereum/client-go:alpine",
							Args:  []string{"--config", "/etc/geth/cfg/config.toml", "--datadir", "/etc/geth/.ethereum"},
							Ports: []corev1.ContainerPort{
								{
									Name:          "geth-p2p-listen",
									ContainerPort: 30303,
								},
								{
									Name:          "geth-p2p-discovery",
									ContainerPort: 30304,
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
