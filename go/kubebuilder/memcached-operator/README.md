# Memcached Go Operator

## Overview

This Memcached operator is a simple example operator using the [Operator SDK][operator_sdk] CLI tool and controller-runtime library API.
For more detailed information on project creation, please refer [Quickstart][quickstart].

## Prerequisites

- [go][go_tool] version v1.13+.
- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.11.3+
- [kustomize][kustomize] v3.1.0+
- [operator-sdk][operator_install]
- Access to a Kubernetes v1.11.3+ cluster

## Getting Started

### Cloning the repository

Checkout this Memcached Operator repository

```
$ mkdir -p $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/go/kubebuilder/memcached-operator
```
### Pulling the dependencies

Run the following command

```
$ go mod tidy
```

### Building the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ make install
$ export IMG=quay.io/example-inc/memcached-operator:v0.0.1
$ make docker-build $IMG
$ docker push $IMG
$ make deploy $IMG
```

**NOTE** The `quay.io/example-inc/memcached-operator:v0.0.1` is an example. You should build and push the image for your repository.

Please verify expected result.

```shell
$ kubectl get all -n memcached-operator-system 
NAME                                                         READY   STATUS    RESTARTS   AGE
pod/memcached-operator-controller-manager-864f7c75d4-7cf47   2/2     Running   0          118s
pod/memcached-sample-68b656fbd4-2pnl8                        1/1     Running   0          53s
pod/memcached-sample-68b656fbd4-464xk                        1/1     Running   0          53s
pod/memcached-sample-68b656fbd4-gzz5l                        1/1     Running   0          53s

NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/memcached-operator-controller-manager-metrics-service   ClusterIP   10.96.171.209   <none>        8443/TCP   118s

NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/memcached-operator-controller-manager   1/1     1            1           118s
deployment.apps/memcached-sample                        3/3     3            3           53s

NAME                                                               DESIRED   CURRENT   READY   AGE
replicaset.apps/memcached-operator-controller-manager-864f7c75d4   1         1         1       118s
```

### Uninstalling

To uninstall all that was performed in the above step run `make uninstall`.

### Troubleshooting

Use the following command to check the operator logs.

```shell
$ kubectl logs deployment.apps/memcached-operator-controller-manager -n  memcached-operator-system -c manager

```

[go_tool]: https://golang.org/dl/
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[kustomize]: https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
[quickstart]: https://github.com/operator-framework/operator-sdk/blob/master/website/content/en/docs/kubebuilder/quickstart.md#implement-the-controller
