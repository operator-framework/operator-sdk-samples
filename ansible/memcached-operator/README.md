# Memcached Operator

## Overview

This Memcached operator is a simple example operator for the [Operator SDK][operator_sdk]. It includes:
  * A Custom Resource Definition for `Memcached` resources
  * An Ansible-based controller to respond to `Memcached` resources
  * Molecule-based tests for the Ansible role.

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator_sdk][operator_install]
- Access to a Kubernetes v1.14.1+ cluster

## Getting Started

### Cloning the repository

Checkout this Memcached Operator repository

```
git clone https://github.com/operator-framework/operator-sdk-samples.git
cd operator-sdk-samples/ansible/memcached-operator
```

### Building the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```sh
export IMG=quay.io/example-inc/memcached-operator:v0.0.1
docker push IMG=$IMG
```

**NOTE** To allow the cluster pull the image the repository needs to be set as public or you must configure an image pull secret.

### Run the operator

Deploy the project to the cluster. Set `IMG` with `make deploy` to use the image you just pushed:

### Create a `Memcached` resource

Apply the sample Custom Resource:

```sh
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml -n memcached-operator-system
```

Run the following command to verify that the installation was successful:

```console
$ kubectl get all -n memcached-operator-system

NAME                                                        READY   STATUS    RESTARTS   AGE
pod/memcached-operator-controller-manager-f896cd75b-v5jqc   2/2     Running   0          19s
pod/memcached-sample-memcached-6456bdd5fc-hwgnd             1/1     Running   0          12s

NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/memcached-operator-controller-manager-metrics-service   ClusterIP   10.102.107.68   <none>        8443/TCP   19s

NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/memcached-operator-controller-manager   1/1     1            1           19s
deployment.apps/memcached-sample-memcached              1/1     1            1           12s

NAME                                                              DESIRED   CURRENT   READY   AGE
replicaset.apps/memcached-operator-controller-manager-f896cd75b   1         1         1       19s
replicaset.apps/memcached-sample-memcached-6456bdd5fc             1         1         1       12s
```

### Cleanup

To leave the operator, but remove the memcached sample pods, delete the
CR.

```sh
kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml -n memcached-operator-system
```

To clean up everything:

```sh
make undeploy
```
### Troubleshooting

Run the following command to check the operator logs.

```sh
kubectl logs deployment.apps/memcached-operator-controller-manager -n memcached-operator-system -c manager
```

### Extras

This project was created by using the [gen-ansible-memcached.sh][gen-ansible-memcached.sh] script.

For mor information see the [Ansible-based operator docs][ansible-docs].

[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[ansible-docs]: https://sdk.operatorframework.io/docs/building-operators/ansible/
[gen-ansible-memcached.sh]: .generate/gen-helm-memcached.sh
