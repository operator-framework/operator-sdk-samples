# Memcached Helm Operator

## Overview

This Memcached operator is a simple example operator based in Helm built with the [Operator SDK][operator_sdk].

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.11.3+
- [operator_sdk][operator_install]
- Access to a Kubernetes v1.11.3+ cluster

## Getting Started

### Cloning the repository

Checkout this Memcached Operator repository

```
$ mkdir operator-framework
$ cd operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/helm/memcached-operator
```

### Building the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/memcached-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

**NOTE** The `quay.io/example-inc/memcached-operator` is an example. You should build and push the image for your repository. 

### Using the image

```
# Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator|g' deploy/operator.yaml
```

### Installing

Run `make install` to install the operator. Check that the operator is running in the cluster, also check that the example Memcached service was deployed.

Following the expected result. 

```shell
$ kubectl get all -n helm-memcached
NAME                                      READY   STATUS    RESTARTS   AGE
pod/example-memcached-84dc867dc-b28wc     1/1     Running   0          7s
pod/example-memcached-84dc867dc-vrxd8     1/1     Running   0          7s
pod/example-memcached-84dc867dc-vvb29     1/1     Running   0          7s
pod/memcached-operator-5b45959c8b-sx9x4   1/1     Running   0          13s

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/example-memcached            ClusterIP   10.107.76.22    <none>        80/TCP              7s
service/memcached-operator-metrics   ClusterIP   10.99.126.244   <none>        8686/TCP,8383/TCP   7s

NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/example-memcached    3/3     3            3           7s
deployment.apps/memcached-operator   1/1     1            1           13s

NAME                                            DESIRED   CURRENT   READY   AGE
replicaset.apps/example-memcached-84dc867dc     3         3         3       7s
replicaset.apps/memcached-operator-5b45959c8b   1         1         1       13s
```

### Uninstalling 

To uninstall all that was performed in the above step run `make uninstall`.

### Troubleshooting

Run the following command to check the operator logs. 

```shell
kubectl logs deployment.apps/memcached-operator -n helm-memcached
```

**NOTE** To have further information about how to develop Helm operators with [Operator-SDK][operator_sdk] check the [Helm User Guide for Operator-SDK][helm_guide]

### Extras

Mote that this project was created by using the following command which means that it is using the [stable/memcached][stable/memcached]

```shell
operator-sdk new memcached-operator --api-version=cache.example.com/v1alpha1 --kind=Memcached --type=helm --helm-chart=stable/memcached
```

[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
[helm_guide]: https://github.com/operator-framework/operator-sdk/blob/master/doc/helm/user-guide.md
[stable/memcached]: https://github.com/helm/charts/tree/master/stable/memcached