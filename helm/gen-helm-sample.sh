#!/usr/bin/env bash

#  Copyright 2019 The Operator-SDK Authors
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

set -o errexit
set -o pipefail

# Turn colors in this script off by setting the NO_COLOR variable in your
# environment to any value:
#
# $ NO_COLOR=1 test.sh
NO_COLOR=${NO_COLOR:-""}
if [ -z "$NO_COLOR" ]; then
  header=$'\e[1;33m'
  reset=$'\e[0m'
else
  header=''
  reset=''
fi

function header_text {
  echo "$header$*$reset"
}

function gen_helm_sample {
  header_text "starting to generate the sample ..."
  dir=$(pwd)
  operIMG="quay.io/example-inc/memcached-operator:v0.0.1"
  bundleIMG="quay.io/example-inc/memcached-operator-bundle:v0.0.1"

  header_text "removing memcached-operator ..."
  rm -rf "$dir/memcached-operator"

  header_text "creating memcached-operator  ..."
  mkdir -p "$dir/memcached-operator"
  cd "$dir/memcached-operator"
  operator-sdk init --plugins=helm --domain=example.com
  operator-sdk create api --version=v1alpha1 --group=cache --kind=Memcached  --helm-chart=stable/memcached

  header_text "customizing sample project ..."

  header_text "updating config/samples/cache_v1alpha1_memcached.yaml ..."
  sed -i".bak" -E -e 's/AntiAffinity: hard/AntiAffinity: soft/g' config/samples/cache_v1alpha1_memcached.yaml; rm -f config/samples/cache_v1alpha1_memcached.yaml.bak

  header_text "adding policy rbac roles ..."
  sed -i".bak" -E -e '/kubebuilder/d' config/rbac/role.yaml; rm -f config/rbac/role.yaml.bak
  cat ../policy-role.yaml >> config/rbac/role.yaml

  header_text "enabling prometheus metrics..."
  sed -i".bak" -E -e 's/(#- ..\/prometheus)/- ..\/prometheus/g' config/default/kustomization.yaml; rm -f config/default/kustomization.yaml.bak

  header_text "integrating with OLM ..."
  header_text "customize bundle target into Makefile to set --interactive=false ..."
  sed -i".bak" -E -e 's/operator-sdk generate kustomize manifests/operator-sdk generate kustomize manifests --interactive=false/g' Makefile; rm -f Makefile.bak


  header_text "generating bundle and building the image $bundleIMG ..."
  make bundle IMG=$bundleIMG
  make bundle-build BUNDLE_IMG=$bundleIMG

  header_text "adding packagemanifests target to the Makefile ..."
  cat ../packagemanifests-target.txt >> Makefile

  header_text "generating packagemanifests for the image $operIMG..."
  make packagemanifests IMG=$operIMG
}

gen_helm_sample
