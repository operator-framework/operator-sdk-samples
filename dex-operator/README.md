# Dex Operator

## Overview
This Dex operator is an operator built with the [Operator SDK][operator_sdk]. Currently the operator supports scaling the numbers of pods on the mock connector and future work involves allowing the connectors to be configured using the operator CRD.

## Quick Start
This quick start guide walks through the process of building the dex-operator.

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
$ go install github.com/operator-framework/operator-sdk/commands/operator-sdk
```

### Initial Setup
Checkout this Dex Operator repository
```
$ mkdir $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk-samples.git
$ cd operator-sdk-samples/dex-operator
```
Vendor the dependencies
```
$ dep ensure
```

### Build the operator
Build the Memcached operator image and push it to a public registry, such as quay.io
```
$ export IMAGE=quay.io/example-inc/dex-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[operator_sdk]:https://github.com/operator-framework/operator-sdk
