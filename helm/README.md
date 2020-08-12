# Memcached Helm Operator

## Overview

This Memcached operator is a simple example of the Operator SDK Helm-based operator. It is based on the [`stable/memcached` chart][stable/memcached] .

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator SDK][operator_install]
- Access to a Kubernetes v1.16.0+ cluster.

## Getting Started

### Clone the repository

Checkout this Memcached operator repository

```
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/helm/memcached-operator
```

### Build the operator image

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
export IMAGE=quay.io/example-inc/memcached-operator:v0.0.1
make docker-build docker-push IMG=$IMAGE
```

**NOTE** To allow the cluster pull the image the repository needs to be set as public or you must configure an image pull secret.


### Run the operator

Deploy the project to the cluster. Set `IMG` with `make deploy` to use the image you just pushed:

```sh
make deploy IMG=$IMAGE
```

### Create a sample custom resource

Create a sample CR:

```sh
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml -n memcached-operator-system
```

Run the following command to verify that the installation was successful:

```console
$ kubectl get all -n memcached-operator-system
NAME                                                        READY   STATUS    RESTARTS   AGE
pod/memcached-operator-controller-manager-d54b5fb78-ltwqs   2/2     Running   0          2m16s
pod/memcached-sample-0                                      1/1     Running   0          96s
pod/memcached-sample-1                                      1/1     Running   0          82s
pod/memcached-sample-2                                      1/1     Running   0          72s

NAME                                                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/memcached-operator-controller-manager-metrics-service   ClusterIP   10.107.115.48   <none>        8443/TCP    2m16s
service/memcached-sample                                        ClusterIP   None            <none>        11211/TCP   96s

NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/memcached-operator-controller-manager   1/1     1            1           2m16s

NAME                                                              DESIRED   CURRENT   READY   AGE
replicaset.apps/memcached-operator-controller-manager-d54b5fb78   1         1         1       2m16s

NAME                                READY   AGE
statefulset.apps/memcached-sample   3/3     96s
```

### Clean up

Delete the CR to uninstall the release:
`
```sh
kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml -n memcached-operator-system
````

Use `make uninstall` and `make undeploy` to uninstall the operator and its CRDs:

```sh
make uninstall
make undeploy
```

### Troubleshooting

Run the following command to check the operator logs.

```sh
kubectl logs deployment.apps/memcached-operator-controller-manager -n memcached-operator-system -c manager
```

### Extras

This project was created by using the [gen-helm-memcached.sh][gen-helm-memcached.sh] script , which means that it is using the official [stable/memcached][stable/memcached] helm chart.

Note that you must have Helm installed locally and add the stable helm charts repo to it to use the `stable/memcached` Helm charts. See the [Helm Quickstart guide][helm-quick] for installation instructions. Also, you can start with Helm-based Operators with SDK by checking its [quickstart][helm_guide].

[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[helm_guide]: https://sdk.operatorframework.io/docs/helm/quickstart/
[stable/memcached]: https://github.com/helm/charts/tree/master/stable/memcached
[helm-quick]: https://helm.sh/docs/intro/quickstart/
[gen-helm-memcached.sh]: .generate/gen-helm-memcached.sh
