# Memcached Operator

## Overview

This Memcached operator is a simple example operator for the [Operator SDK][operator_sdk] and includes some basic end-to-end tests.

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator_sdk][operator_install] v1.0.0+
- Access to a Kubernetes v1.14.1+ cluster

## Getting Started

### Cloning the repository

Checkout this Memcached Operator repository

```
$ mkdir operator-framework
$ cd operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/ansible/memcached-operator
```

## Building and Pushing the Project Image

To build and push your image to your repository :
```
$ make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```

Note: To allow the cluster pull the image the repository needs to be set as public.

## Applying the CRDs into the cluster:

To apply the Memcached Kind(CRD):
```
$ make install
```

## Applying the CR’s into the cluster:

To create instances (CR’s) of the Memcached Kind (CRD) in the same namespaced of the operator:
```
$ kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml -n memcached-operator-system
```

## Running it on Cluster
Deploy the project to the cluster:

```
$ make deploy IMG=<some-registry>/<project-name>:tag
```

Following the expected result.

```shell
 $ kubectl get all -n memcached-operator-system
NAME                                                         READY   STATUS    RESTARTS   AGE
pod/memcached-operator-controller-manager-7dbcd676f9-s9nrz   2/2     Running   0          113s
pod/memcached-sample-memcached-6456bdd5fc-fdbfg              1/1     Running   0          67s
pod/memcached-sample-memcached-6456bdd5fc-p97h8              1/1     Running   0          67s
pod/memcached-sample-memcached-6456bdd5fc-q4wbb              1/1     Running   0          67s

NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/memcached-operator-controller-manager-metrics-service   ClusterIP   10.96.153.103   <none>        8443/TCP   4h52m

NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/memcached-operator-controller-manager   1/1     1            1           4h52m
deployment.apps/memcached-sample-memcached              3/3     3            3           67s

NAME                                                               DESIRED   CURRENT   READY   AGE
replicaset.apps/memcached-operator-controller-manager-5b7f656f48   0         0         0       4h52m
replicaset.apps/memcached-operator-controller-manager-7dbcd676f9   1         1         1       113s
replicaset.apps/memcached-operator-controller-manager-7dbdb5b9ff   0         0         0       52m
replicaset.apps/memcached-sample-memcached-6456bdd5fc              3         3         3       67s
```

### Uninstalling

To uninstall all that was performed in the above step run `make uninstall`.

### Troubleshooting

Use the following command to check the operator logs.

```
$ kubectl logs deployment.apps/memcached-operator-controller-manager -n memcached-operator-system -c manager
```

**NOTE** To have further information about how to develop Ansible operators with [Operator-SDK][operator_sdk] check the [Ansible docs][ansible-docs].


[python]: https://www.python.org/
[ansible]: https://www.ansible.com/
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[ansible-docs]: https://sdk.operatorframework.io/docs/docs/building-operators/ansible/
