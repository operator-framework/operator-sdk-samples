package e2eutil

import (
	goctx "context"
	"testing"
	"time"

	eopapi "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type acceptFunc func(*eopapi.EtcdCluster) bool

func EtcdWaitUntilSizeReached(t *testing.T, f *framework.Framework, size, retries int, cl *eopapi.EtcdCluster) ([]string, error) {
	return etcdWaitSizeReachedWithAccept(t, f, size, retries, cl)
}

func etcdWaitSizeReachedWithAccept(t *testing.T, f *framework.Framework, size, retries int, cl *eopapi.EtcdCluster, accepts ...acceptFunc) ([]string, error) {
	var names []string
	err := wait.PollImmediate(retryInterval, time.Duration(retries)*retryInterval, func() (done bool, err error) {
		currCluster := &eopapi.EtcdCluster{}
		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: cl.Namespace, Name: cl.Name}, currCluster)
		if err != nil {
			return false, err
		}

		for _, accept := range accepts {
			if !accept(currCluster) {
				return false, nil
			}
		}

		names = currCluster.Status.Members.Ready
		LogfWithTimestamp(t, "waiting size (%d), healthy etcd members: names (%v)", size, names)
		if len(names) != size {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}
