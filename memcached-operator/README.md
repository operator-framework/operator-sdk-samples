# Memcached Ansible Operator

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
$ mkdir $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/memcached-operator
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
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator|g' deploy/operator.yaml
```

**NOTE** The `quay.io/example-inc/memcached-operator` is an example. You should build and push the image for your repository. 

### Installing

Run `make install` to install the operator. Check that the operator is running in the cluster, also check that the example Memcached service was deployed.

### Uninstalling 

To uninstall all that was performed in the above step run `make uninstall`.

[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go_tool]:https://golang.org/dl/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]:https://docs.docker.com/install/
[operator_sdk]:https://github.com/operator-framework/operator-sdk
[operator_install]:https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md

### Run Tests

Run `make test-e2e` to run integration tests. This command will execute many times the same e2e tests. Its purposes is just to illustrate few options.