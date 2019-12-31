# Memcached Go Operator

## Overview

This Memcached operator is a simple example operator for the [Operator SDK][operator_sdk] and includes some basic end-to-end tests.

## Prerequisites

- [dep][dep_tool] version v0.5.0+.
- [go][go_tool] version v1.12+.
- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.11.3+
- [operator-sdk][operator_install]
- Access to a Kubernetes v1.11.3+ cluster

## Getting Started

### Cloning the repository

Checkout this Memcached Operator repository

```
$ mkdir -p $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/go/memcached-operator
```
### Pulling the dependencies

Run the following command

```
$ go mod tidy
```

### Building the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/memcached-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

### Using the image

```
# Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator:v0.0.1|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator:v0.0.1|g' deploy/operator.yaml
```

**NOTE** The `quay.io/example-inc/memcached-operator:v0.0.1` is an example. You should build and push the image for your repository.

### Installing

Run `make install` to install the operator. Check that the operator is running in the cluster, also check that the example Memcached service was deployed.

Following the expected result.

```shell
$ kubectl get all -n memcached 
NAME                                      READY   STATUS    RESTARTS   AGE
pod/example-memcached-7c4df9b7b4-lzd6j    1/1     Running   0          64s
pod/example-memcached-7c4df9b7b4-wbtkz    1/1     Running   0          64s
pod/example-memcached-7c4df9b7b4-wt6jb    1/1     Running   0          64s
pod/memcached-operator-56f54d84bf-zrtfv   1/1     Running   0          69s

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/example-memcached            ClusterIP   10.108.124.47   <none>        11211/TCP           63s
service/memcached-operator-metrics   ClusterIP   10.108.67.82    <none>        8383/TCP,8686/TCP   66s

NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/example-memcached    3/3     3            3           65s
deployment.apps/memcached-operator   1/1     1            1           70s

NAME                                            DESIRED   CURRENT   READY   AGE
replicaset.apps/example-memcached-7c4df9b7b4    3         3         3       65s
replicaset.apps/memcached-operator-56f54d84bf   1         1         1       70s
```

### Uninstalling

To uninstall all that was performed in the above step run `make uninstall`.

### Troubleshooting

Use the following command to check the operator logs.

```shell
kubectl logs deployment.apps/memcached-operator -n memcached
```

### Running Tests

Run `make test-e2e` to run the integration e2e tests with different options. For
more information see the [writing e2e tests](https://github.com/operator-framework/operator-sdk/blob/master/doc/test-framework/writing-e2e-tests.md) guide.

[dep_tool]: https://golang.github.io/dep/docs/installation.html
[go_tool]: https://golang.org/dl/
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
