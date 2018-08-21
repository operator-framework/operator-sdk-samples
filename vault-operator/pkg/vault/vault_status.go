package vault

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"

	vaultapi "github.com/hashicorp/vault/api"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func updateVaultStatus(vr *api.VaultService, status *api.VaultServiceStatus) error {
	// don't update the status if there aren't any changes.
	if reflect.DeepEqual(vr.Status, *status) {
		return nil
	}
	vr.Status = *status
	return sdk.Update(vr)
}

// getVaultStatus retrieves the status of the vault cluster for the given Custom Resource "vr",
// and it only succeeds if all of the nodes from vault cluster are reachable.
func getVaultStatus(vr *api.VaultService) (*api.VaultServiceStatus, error) {
	pods := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
	sel := LabelsForVault(vr.Name)
	opt := &metav1.ListOptions{LabelSelector: labels.SelectorFromSet(sel).String()}
	err := sdk.List(vr.GetNamespace(), pods, sdk.WithListOptions(opt))
	if err != nil {
		return nil, fmt.Errorf("failed to get vault's pods: %v", err)
	}

	tc, err := vaultTLSFromSecret(vr)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS config for vault client: %v", err)
	}

	var (
		initialized bool
		active      string
		standby     []string
		sealed      []string
		updated     []string
	)
	for _, p := range pods.Items {
		// If a pod is terminating, then we can't access the corresponding vault node's status.
		// so we break from here and return an error.
		if p.Status.Phase != v1.PodRunning || p.DeletionTimestamp != nil {
			return nil, errors.New("vault pod is terminating")
		}

		vapi, err := NewVaultClient(PodDNSName(p), "8200", tc)
		if err != nil {
			return nil, fmt.Errorf("failed creating client for the vault pod (%s/%s): %v", vr.GetNamespace(), p.GetName(), err)
		}

		hr, err := vapi.Sys().Health()
		if err != nil {
			return nil, fmt.Errorf("failed requesting health info for the vault pod (%s/%s): %v", vr.GetNamespace(), p.GetName(), err)
		}

		if isVaultVersionMatch(p.Spec, vr.Spec) {
			updated = append(updated, p.GetName())
		}

		if hr.Initialized && !hr.Sealed && !hr.Standby {
			active = p.GetName()
		}
		if hr.Initialized && !hr.Sealed && hr.Standby {
			standby = append(standby, p.GetName())
		}
		if hr.Sealed {
			sealed = append(sealed, p.GetName())
		}
		if hr.Initialized {
			initialized = true
		}
	}

	return &api.VaultServiceStatus{
		Phase:       api.ClusterPhaseRunning,
		Initialized: initialized,
		ServiceName: vr.GetName(),
		ClientPort:  vaultClientPort,
		VaultStatus: api.VaultStatus{
			Active:  active,
			Standby: standby,
			Sealed:  sealed,
		},
		UpdatedNodes: updated,
	}, nil
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", hostname, port)
	cfg.Address = podURL
	cfg.ConfigureTLS(tlsConfig)
	return vaultapi.NewClient(cfg)
}

// VaultTLSFromSecret reads Vault CR's TLS secret and converts it into a vault client's TLS config struct.
func vaultTLSFromSecret(vr *api.VaultService) (*vaultapi.TLSConfig, error) {
	cs := vr.Spec.TLS.Static.ClientSecret
	se := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cs,
			Namespace: vr.GetNamespace(),
		},
	}
	err := sdk.Get(se)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: failed to get secret (%s): %v", cs, err)
	}

	// Read the secret and write ca.crt to a temporary file
	caCertData := se.Data[api.CATLSCertName]
	f, err := ioutil.TempFile("", api.CATLSCertName)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: create temp file failed: %v", err)
	}
	defer f.Close()

	_, err = f.Write(caCertData)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: write ca cert file failed: %v", err)
	}
	if err = f.Sync(); err != nil {
		return nil, fmt.Errorf("read client tls failed: sync ca cert file failed: %v", err)
	}
	return &vaultapi.TLSConfig{CACert: f.Name()}, nil
}

// PodDNSName constructs the dns name on which a pod can be addressed
func PodDNSName(p v1.Pod) string {
	podIP := strings.Replace(p.Status.PodIP, ".", "-", -1)
	return fmt.Sprintf("%s.%s.pod", podIP, p.Namespace)
}
