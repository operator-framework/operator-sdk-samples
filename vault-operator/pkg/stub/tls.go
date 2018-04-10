package stub

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/coreos-inc/vault-operator/pkg/util/k8sutil"
	"github.com/coreos-inc/vault-operator/pkg/util/tlsutil"
)

// prepareEtcdTLSSecrets creates three etcd TLS secrets (client, server, peer) containing TLS assets.
// Currently we self-generate the CA, and use the self generated CA to sign all the TLS certs.
func (v *Vaults) prepareEtcdTLSSecrets(vr *api.VaultService) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("prepare TLS secrets failed: %v", err)
		}
	}()

	// TODO: use secrets informer
	_, err = v.kubecli.CoreV1().Secrets(vr.Namespace).Get(k8sutil.EtcdClientTLSSecretName(vr.Name), metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// TODO: optional user pass-in CA.
	caKey, caCrt, err := newCACert()
	if err != nil {
		return err
	}

	se, err := newEtcdClientTLSSecret(vr, caKey, caCrt)
	if err != nil {
		return err
	}
	k8sutil.AddOwnerRefToObject(se, k8sutil.AsOwner(vr))
	_, err = v.kubecli.CoreV1().Secrets(vr.Namespace).Create(se)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	se, err = newEtcdServerTLSSecret(vr, caKey, caCrt)
	if err != nil {
		return err
	}
	k8sutil.AddOwnerRefToObject(se, k8sutil.AsOwner(vr))
	_, err = v.kubecli.CoreV1().Secrets(vr.Namespace).Create(se)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	se, err = newEtcdPeerTLSSecret(vr, caKey, caCrt)
	if err != nil {
		return err
	}
	k8sutil.AddOwnerRefToObject(se, k8sutil.AsOwner(vr))
	_, err = v.kubecli.CoreV1().Secrets(vr.Namespace).Create(se)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func newCACert() (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := tlsutil.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	config := tlsutil.CertConfig{
		CommonName:   "vault operator CA",
		Organization: orgForTLSCert,
	}

	cert, err := tlsutil.NewSelfSignedCACertificate(config, key)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, err
}
