# Vault Operator

## Overview 

This Vault operator is a re-implementation of the [Vault operator][vault_operator] using the [operator-sdk][operator_sdk] tools and APIs. The SDK CLI `operator-sdk` generates the project layout and controls the development life cycle. In addition, this implementation replaces the use of [client-go][client_go] with the SDK APIs to watch, query, and mutate Kubernetes resources.

## Quick Start

The quick start guide walks through the process of building the Vault operator image using the SDK CLI, setting up the RBAC, deploying operators, and creating a vault cluster.

### Prerequisites

- [dep][dep_tool] version v0.5.0+.
- [go][go_tool] version v1.10+.
- [docker][docker_tool] version 17.03+.
- [kubectl][kubectl_tool] version v1.9.0+.
- Access to a kubernetes v.1.9.0+ cluster.

**Note**: This guide uses [minikube][minikube_tool] version v0.25.0+ as the local kubernetes cluster and quay.io for the public registry.

### Install the Operator SDK CLI

First, checkout and install the operator-sdk CLI:

```sh
$ cd $GOPATH/src/github.com/operator-framework/operator-sdk
$ git checkout master
$ dep ensure
$ go install github.com/operator-framework/operator-sdk/commands/operator-sdk
```

### Initial Setup

Checkout this Vault Operator repository:

```sh
$ mkdir $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/vault-operator
```

Vendor the dependencies:

```sh
$ dep ensure
```

### Build and run the operator

Build the Vault operator image and push it to a public registry such as quay.io:

```sh
$ export IMAGE=quay.io/example/vault-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

Setup RBAC for the Vault operator and its related resources:

```sh
$ kubectl create -f deploy/rbac.yaml
```

Deploy the etcd-operator first because the Vault operator depends on it for provisioning an etcd cluster as the  storage backend of a Vault cluster:

```sh
$ kubectl create -f deploy/etcd-operator.yaml
```

Deploy the Vault CRD:

```sh
$ kubectl create -f deploy/crd.yaml
```

Deploy the Vault operator:

```sh
$ kubectl create -f deploy/operator.yaml
```
### Deploying a Vault cluster

Create a Vault cluster:

```sh
$ kubectl create -f deploy/cr.yaml
```

Verify that the Vault cluster is up:

```sh
$ kubectl get pods -l app=vault,vault_cluster=example
NAME                       READY     STATUS    RESTARTS   AGE
example-654658f5fc-2wdlq   1/2       Running   0          1m
example-654658f5fc-7ztzf   1/2       Running   0          1m
```

### Vault Guide

Once the vault cluster is up, see the [Vault Usage Guide][guide] from the original Vault operator repository on how to initialize, unseal, and interact with the vault cluster.

**Note** The [Vault Usage Guide][guide] uses the short name `vault` for the kind `VaultService`. However, we have not register a short name for this vault Custom Resource Definition (CRD). As a workaround when use a command from [Vault Usage Guide][guide] that has the `vault` keyword to access a vault Custom Resource(CR), replace it with the keyword `vaultservice` instead.

For example:

`kubectl -n default get vault example ...` -> `kubectl -n default get vaultservice example ...`

[client_go]:https://github.com/kubernetes/client-go
[vault_operator]:https://github.com/coreos/vault-operator
[operator_sdk]:https://github.com/operator-framework/operator-sdk
[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[minikube_tool]:https://github.com/kubernetes/minikube#installation
[guide]:https://github.com/coreos/vault-operator/blob/master/doc/user/vault.md 
