# Memcached Helm Operator

## Overview

This Memcached operator is a simple example of the Operator SDK Helm-based operator. It is based on the [`stable/memcached` chart][stable/memcached] .

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator SDK][operator_install]
- Access to a Kubernetes v1.14.1+ cluster

## Getting Started

### Cloning the repository

Checkout this Memcached operator repository

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

**NOTE** The `quay.io/example-inc/memcached-operator:v0.0.1` is an example. You should build and push the image for your repository.

### Using the image

Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)

```
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator:v0.0.1|g' deploy/operator.yaml
```

**Note**
If you are performing these steps on OSX, use the following `sed` command instead:
```
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator:v0.0.1|g' deploy/operator.yaml
```

### Installing

Run `make install` to install the operator. Check that the operator is running in the cluster, also check that the example Memcached service was deployed.

Run the following command to verify that the installation was successful:

```shell
$ kubectl get all -n helm-memcached -o wide
NAME                                      READY   STATUS    RESTARTS   AGE   IP           NODE       NOMINATED NODE   READINESS GATES
pod/example-memcached-0                   1/1     Running   0          37s   172.17.0.5   minikube   <none>           <none>
pod/example-memcached-1                   1/1     Running   0          19s   172.17.0.6   minikube   <none>           <none>
pod/example-memcached-2                   1/1     Running   0          12s   172.17.0.7   minikube   <none>           <none>
pod/memcached-operator-55d98c7cf8-x6x9p   1/1     Running   0          52s   172.17.0.4   minikube   <none>           <none>

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE   SELECTOR
service/example-memcached            ClusterIP   None            <none>        11211/TCP           37s   app=example-memcached
service/memcached-operator-metrics   ClusterIP   10.96.212.206   <none>        8686/TCP,8383/TCP   38s   name=memcached-operator

NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS           IMAGES                            SELECTOR
deployment.apps/memcached-operator   1/1     1            1           52s   memcached-operator   cmacedo/memcached-operator:test   name=memcached-operator

NAME                                            DESIRED   CURRENT   READY   AGE   CONTAINERS           IMAGES                            SELECTOR
replicaset.apps/memcached-operator-55d98c7cf8   1         1         1       52s   memcached-operator   cmacedo/memcached-operator:test   name=memcached-operator,pod-template-hash=55d98c7cf8

NAME                                 READY   AGE   CONTAINERS          IMAGES
statefulset.apps/example-memcached   3/3     37s   example-memcached   memcached:1.5.12-alpine
```

### Uninstalling 

Run `make uninstall` to uninstall all that was performed in the above step.

### Troubleshooting

Run the following command to check the operator logs. 

```shell
kubectl logs deployment.apps/memcached-operator -n helm-memcached
```

**NOTE** For further information about how to develop Helm operators with Operator SDK, read the [Helm User Guide for Operator SDK][helm_guide]

### Extras

This project was created by using the following command, which means that it is using the official [stable/memcached][stable/memcached] helm chart.

```shell
operator-sdk new memcached-operator --api-version=cache.example.com/v1alpha1 --kind=Memcached --type=helm --helm-chart=stable/memcached
```

Note that you must have Helm installed locally and add the stable helm charts repo to it to use the `stable/memcached` Helm charts. See the [Helm Quick Start guide][helm-quick] for installation instructions.

[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_install]: https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md
[helm_guide]: https://github.com/operator-framework/operator-sdk/blob/master/doc/helm/user-guide.md
[stable/memcached]: https://github.com/helm/charts/tree/master/stable/memcached
[helm-quick]: https://helm.sh/docs/intro/quickstart/
