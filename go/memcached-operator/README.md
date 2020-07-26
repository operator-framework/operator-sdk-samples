# Memcached Go Operator

## Overview

This Memcached operator is a simple example operator for the [Operator SDK][operator_sdk] and includes some basic end-to-end tests.

## Prerequisites
Ensure that you have installed [operator-sdk][operator_install] version >= 0.19+ as its pre-requirements. You also
will requires have access to a Kubernetes v1.14.1+ cluster.


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
$ make docker-build docker-push IMG=$IMAGE
```

### Running operator

1. Run `make install` to install your CRD's. 
2. Run `kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml -n memcached-operator-system` to apply the CR
3. Run `make deploy IMG=$IMAGE` to deploy your project   

Following the expected result.

```shell
$ kubectl get all -n memcached-operator-system 
NAME                                                         READY   STATUS    RESTARTS   AGE
pod/memcached-operator-controller-manager-77fb5d86b9-bfdhs   2/2     Running   0          15s
pod/memcached-sample-9b765dfc8-4qzfg                         1/1     Running   0          2s
pod/memcached-sample-9b765dfc8-h9vk8                         1/1     Running   0          2s
pod/memcached-sample-9b765dfc8-q9rdv                         1/1     Running   0          2s

NAME                                                            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/memcached-operator-controller-manager-metrics-service   ClusterIP   10.100.227.5   <none>        8443/TCP   15s

NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/memcached-operator-controller-manager   1/1     1            1           15s
deployment.apps/memcached-sample                        3/3     3            3           2s

NAME                                                               DESIRED   CURRENT   READY   AGE
replicaset.apps/memcached-operator-controller-manager-77fb5d86b9   1         1         1       15s
replicaset.apps/memcached-sample-9b765dfc8                         3         3         3       2s
```

### Uninstalling

To uninstall all that was performed in the above step run `make undeploy`. 

### Troubleshooting

Use the following command to check the operator logs.

```shell
kubectl logs deployment.apps/memcached-operator-controller-manager -n memcached-operator-system -c manager
```

[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
