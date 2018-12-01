# Vault Operator

## Overview

This Vault operator is a re-implementation of the [Vault operator][vault_operator] using the [operator-sdk][operator_sdk] tools and APIs. The SDK CLI `operator-sdk` generates the project layout and controls the development life cycle. In addition, this implementation replaces the use of [client-go][client_go] with the SDK APIs to watch, query, and mutate Kubernetes resources.

## Quick Start

The quick start guide walks through the process of building the Vault operator image using the SDK CLI, setting up the RBAC, deploying operators, and creating a vault cluster.

### Prerequisites

- [dep][dep_tool] version v0.5.0+.
- [git][git_tool]
- [go][go_tool] version v1.10+.
- [docker][docker_tool] version 17.03+.
- [kubectl][kubectl_tool] version v1.11.0+.
- Access to a kubernetes v.1.11.0+ cluster.

### Install the Operator SDK CLI

First, checkout and install the operator-sdk CLI:

```sh
$ cd $GOPATH/src/github.com/operator-framework/operator-sdk
$ git checkout master
$ make dep
$ make install
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

Build the Vault operator image:

```sh
$ export IMAGE=quay.io/example/vault-operator:v0.0.1
$ operator-sdk build $IMAGE
```

Insert your built image name into `deploy/operator.yaml`:

```sh
$ sed -i 's|REPLACE_IMAGE|'"$IMAGE"'|g' deploy/operator.yaml
```

**Note**
If you are performing these steps on OSX, use the following command:
```sh
$ sed -i "" 's|REPLACE_IMAGE|'"$IMAGE"'|g' deploy/operator.yaml
```

Set up RBAC roles and role bindings for the Vault operator and its related resources:

```sh
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
```

Deploy the etcd-operator first because the Vault operator depends on it for provisioning an etcd cluster as the storage backend of a Vault cluster:

```sh
$ kubectl create -f deploy/etcd-operator.yaml
```

Deploy the Vault CRD:

```sh
$ kubectl create -f deploy/crds/vault_v1alpha1_vaultservice_crd.yaml
```

Deploy the Vault operator:

```sh
$ kubectl create -f deploy/operator.yaml
```
### Deploying a Vault cluster

Create a Vault cluster:

```sh
$ kubectl create -f deploy/crds/vault_v1alpha1_vaultservice_cr.yaml
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

**Note** The [Vault Usage Guide][guide] uses the short name `vault` for the kind `VaultService`. However, we have not register a short name for this vault Custom Resource Definition (CRD). As a workaround when use a command from [Vault Usage Guide][guide] that has the `vault` keyword to access a vault Custom Resource (CR), replace it with the keyword `vaultservice` instead.

For example:

`kubectl -n default get vault example ...` -> `kubectl -n default get vaultservice example ...`

## Tests
This repo contains some tests that use the operator-sdk's test framework. These tests are based directly on the original vault-operator
tests, and **thus cannot fully complete when run on a local machine and must be run inside a kubernetes cluster instead**. This is a very
specific use case, so it is not handled by the sdk's test framework. However, it is a good example of how to use the framework for
an operator that needs more resources than standard to initialize due to the dependency on etcd. These tests fully initialize a vault
cluster and tear it down when run on a local machine, even though they do fail due to not being able to use the vault-client to
communicate with the vault pods. To run these tests using the specific test init files, modify the vault-operator's spec inside
`deploy/namespaced-init.yaml` to point to your repo containing the vault-operator, and then run this command:

```sh
$ operator-sdk test local ./test/e2e/ --global-manifest deploy/global-init.yaml --namespaced-manifest deploy/namespaced-init.yaml
```

[vault_operator]:https://github.com/coreos/vault-operator
[operator_sdk]:https://github.com/operator-framework/operator-sdk
[client_go]:https://github.com/kubernetes/client-go
[dep_tool]:https://golang.github.io/dep/docs/installation.html
[git_tool]:https://git-scm.com/downloads
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[guide]:https://github.com/coreos/vault-operator/blob/master/doc/user/vault.md
