# Memcached Operator

## Overview

This Memcached operator is a simple example operator for the [Operator SDK][operator_sdk] and includes some basic end-to-end tests.

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator_sdk][operator_install]
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

### Building the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/memcached-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

**NOTE** The `quay.io/example-inc/memcached-operator:v0.0.1` is an example. You should build and push the image for your repository.

### Using the image

```
# Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator:v0.0.1|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator:v0.0.1|g' deploy/operator.yaml
```

### Installing

Run `make install` to install the operator. Check that the operator is running in the cluster, also check that the example Memcached service was deployed.

Following the expected result.

```shell
$ kubectl get all -n memcached

NAME                                              READY   STATUS    RESTARTS   AGE
pod/example-memcached-memcached-b885dcc75-2crw5   1/1     Running   0          22s
pod/example-memcached-memcached-b885dcc75-69mbg   1/1     Running   0          22s
pod/example-memcached-memcached-b885dcc75-92rd7   1/1     Running   0          22s
pod/memcached-operator-df88b85f7-9s98n            2/2     Running   0          36s

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/memcached-operator-metrics   ClusterIP   10.98.192.187   <none>        8383/TCP   31s

NAME                                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/example-memcached-memcached   3/3     3            3           22s
deployment.apps/memcached-operator            1/1     1            1           36s

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/example-memcached-memcached-b885dcc75   3         3         3       22s
replicaset.apps/memcached-operator-df88b85f7            1         1         1       36s
```

### Uninstalling

To uninstall all that was performed in the above step run `make uninstall`.

### Troubleshooting

Use the following command to check the operator logs.

```
kubectl logs deployment.apps/memcached-operator -n memcached
```

**NOTE:** This project is configured with the environment variable `ANSIBLE_DEBUG_LOGS` as `True`, however, note that it is `False` by default.

**NOTE** To have further information about how to develop Ansible operators with [Operator-SDK][operator_sdk] check the [Ansible User Guide for Operator-SDK][ansible-guide]

### Testing the Operator

See [Testing Ansible Operators with Molecule][ansible-test-guide] documentation to know how to use the operator framework features to test it.  

[python]: https://www.python.org/
[ansible]: https://www.ansible.com/
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[ansible-test-guide]: https://sdk.operatorframework.io/docs/ansible/testing-guide/
[ansible-guide]: https://sdk.operatorframework.io/docs/ansible/quickstart/
