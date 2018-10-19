#!/bin/bash

docker build -t galera-ansible-operator:latest .
kubectl create -f deploy/operator.yaml
kubectl create -f deploy/cr.yaml

