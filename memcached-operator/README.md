# Memcached Operator

## Overview

This Memcached operator is a simple example operator for the [Operator SDK][operator_sdk] and includes some basic end-to-end tests.

## Quick Start

This quick start guide walks through the process of building the memcached-operator and running its end-to-end tests.

### Prerequisites

- [dep][dep_tool] version v0.5.0+.
- [go][go_tool] version v1.10+
- [docker][docker_tool] version 17.03+
- Access to a kubernetes v1.9.0+ cluster

### Install the Operator SDK CLI

First, checkout and install the operator-sdk CLI:

```
$ cd $GOPATH/src/github.com/operator-framework/operator-sdk
$ git checkout master // currently, there are no releases that include the test framework, so use the master for now
$ dep ensure
$ make install github.com/operator-framework/operator-sdk/commands/operator-sdk
```

### Initial Setup

Checkout this Memcached Operator repository

```
$ mkdir $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/memcached-operator
```

Vendor the dependencies

```
$ dep ensure
```

### Build the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/memcached-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

### Run the tests

To run the tests in `test/e2e`, first make sure to update the image name from `REPLACE_IMAGE`
to your desired image name in `deploy/operator.yaml`:

```
$ sed "s@REPLACE_IMAGE@$IMAGE@g" -i deploy/operator.yaml
```

Run the tests that reside in `test/e2e`:

```
$ operator-sdk test local ./test/e2e
```

To run go-test with verbose and limit to 2 parallel tests:

```
$ operator-sdk test local ./test/e2e --go-test-flags "-v -parallel=2"
```

To run the tests from an image in the cluster, first build the tests into your image:

```
$ operator-sdk build $IMAGE --enable-tests
$ docker push $IMAGE
```

Then, configure the testing environment:

```
$ export NAMESPACE="my-test"
$ kubectl create -f deploy/crds/cache_v1alpha1_memcached_crd.yaml -n $NAMESPACE
$ kubectl create -f deploy/service_account.yaml -n $NAMESPACE
$ kubectl create -f deploy/role.yaml -n $NAMESPACE
$ kubectl create -f deploy/role_binding.yaml -n $NAMESPACE
```

Finally, run the tests:
```
$ operator-sdk test cluster $IMAGE --namespace $NAMESPACE --service-account memcached-operator
```

[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[operator_sdk]:https://github.com/operator-framework/operator-sdk
