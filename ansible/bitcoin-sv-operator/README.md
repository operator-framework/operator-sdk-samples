# bitcoin-sv-operator

This is an Operator which deploys [Bitcoin Satoshi Vision (SV)](https://github.com/bitcoin-sv/bitcoin-sv) - a Bitcoin Cash full node implementation - on Kubernetes/Openshift. 

This Operator is an [Ansible Operator](https://github.com/water-hole/ansible-operator) which deploys a containerized version of `bitcoind` with RPC and REST capabilities.

This project was generated using [operator-sdk](https://github.com/operator-framework/operator-sdk).

To deploy the operator:
```
$ kubectl create -f deploy/rbac.yaml
$ kubectl create -f deploy/crd.yaml
$ kubectl create -f deploy/operator.yaml
```

To launch an instance of Bitcoin SV:
```
$ kubectl create -f deploy/cr.yaml
```
