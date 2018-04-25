package vault

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// vaultConfigPath is the path that vault pod uses to read config from
	vaultConfigPath = "/run/vault/config/vault.hcl"
	// vaultTLSAssetDir is the dir where vault's server TLS and etcd TLS assets sits
	vaultTLSAssetDir = "/run/vault/tls/"
	// serverTLSCertName is the filename of the vault server cert
	serverTLSCertName = "server.crt"
	// serverTLSKeyName is the filename of the vault server key
	serverTLSKeyName = "server.key"
)

// prepareVaultConfig applies our section into Vault config file.
// - If given user configmap, appends into user provided vault config
//   and creates another configmap "${configMapName}-copy" for it.
// - Otherwise, creates a new configmap "${vaultName}-copy" with our section.
func prepareVaultConfig(vr *api.VaultService) error {
	var cfgData string
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: vr.Namespace,
		},
	}
	if len(vr.Spec.ConfigMapName) != 0 {
		cm.Name = vr.Spec.ConfigMapName
		err := query.Get(cm)
		if err != nil {
			return fmt.Errorf("prepare vault config error: get configmap (%s) failed: %v", vr.Spec.ConfigMapName, err)
		}
		cfgData = cm.Data[filepath.Base(vaultConfigPath)]
	}

	cm.Name = configMapNameForVault(vr)
	cm.Labels = labelsForVault(vr.Name)
	cfgData = newConfigWithDefaultParams(cfgData)
	cfgData = newConfigWithEtcd(cfgData, etcdURLForVault(vr.Name))
	cm.Data = map[string]string{filepath.Base(vaultConfigPath): cfgData}
	addOwnerRefToObject(cm, asOwner(vr))
	err := action.Create(cm)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("prepare vault config error: create new configmap (%s) failed: %v", cm.Name, err)
	}
	return nil
}

const listenerFmt = `
listener "tcp" {
  address     = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file = "%s"
  tls_key_file  = "%s"
}
`

const etcdStorageFmt = `
storage "etcd" {
  address = "%s"
  etcd_api = "v3"
  ha_enabled = "true"
  tls_ca_file = "%s"
  tls_cert_file = "%s"
  tls_key_file = "%s"
  sync = "false"
}
`

// newConfigWithEtcd returns the new config data combining
// original config and new etcd storage section.
func newConfigWithEtcd(data, etcdURL string) string {
	storageSection := fmt.Sprintf(etcdStorageFmt, etcdURL, filepath.Join(vaultTLSAssetDir, "etcd-client-ca.crt"),
		filepath.Join(vaultTLSAssetDir, "etcd-client.crt"), filepath.Join(vaultTLSAssetDir, "etcd-client.key"))
	data = fmt.Sprintf("%s%s", data, storageSection)
	return data
}

// newConfigWithDefaultParams appends to given config data some default params:
// - telemetry setting
// - tcp listener
func newConfigWithDefaultParams(data string) string {
	buf := bytes.NewBufferString(data)
	buf.WriteString(`
telemetry {
	statsd_address = "localhost:9125"
}
`)
	listenerSection := fmt.Sprintf(listenerFmt,
		filepath.Join(vaultTLSAssetDir, serverTLSCertName),
		filepath.Join(vaultTLSAssetDir, serverTLSKeyName))
	buf.WriteString(listenerSection)

	return buf.String()
}

// configMapNameForVault is the configmap name for the given vault.
// If ConfigMapName is given is spec, it will make a new name based on that.
// Otherwise, we will create a default configmap using the Vault's name.
func configMapNameForVault(v *api.VaultService) string {
	n := v.Spec.ConfigMapName
	if len(n) == 0 {
		n = v.Name
	}
	return n + "-copy"
}
