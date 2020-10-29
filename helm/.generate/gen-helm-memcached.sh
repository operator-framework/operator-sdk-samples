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

# Before run this script ensure that you have helm installed locally
# with the stable repo as well. The helm sample will use the memcached chart
# from helm repository.
# To install: https://helm.sh/docs/intro/install/
# To add the repo run `helm repo add stable https://charts.helm.sh/stable`

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

operatorName="memcached-operator"
function header_text {
  echo "$header$*$reset"
}

function gen_helm_sample {

  # When operator-sdk scaffolds Helm projects, it tries to use the discovery API of a Kubernetes
	# cluster to intelligently build the RBAC rules that the operator will require based on the
	# content of the helm chart.
	#
	# Here, we intentionally set KUBECONFIG to a broken value to ensure that operator-sdk will be
	#  unable to reach a real cluster, and thus will generate a default RBAC rule set. This is
	# required to make Helm project generation idempotent because contributors and CI environments
	# can all have slightly different environments that can affect the content of the generated
	# role and cause sanity testing to fail.
  export KUBECONFIG="broken_so_we_generate_static_default_rules"

  local operIMG="quay.io/example-inc/memcached-operator:v0.0.1"
  local bundleIMG="quay.io/example-inc/memcached-operator-bundle:v0.0.1"
  
  header_text "starting to generate the sample ..."

  header_text "removing memcached-operator ..."
  cd ..
  rm -rf $operatorName

  header_text "creating $operatorName"


  header_text "creating memcached-operator  ..."
  mkdir $operatorName
  cd $operatorName
  operator-sdk init --plugins=helm --domain=example.com
  operator-sdk create api --version=v1alpha1 --group=cache --kind=Memcached  --helm-chart=stable/memcached

  header_text "customizing sample project ..."

  header_text "updating config/samples/cache_v1alpha1_memcached.yaml ..."
  sed -i".bak" -E -e 's/AntiAffinity: hard/AntiAffinity: soft/g' config/samples/cache_v1alpha1_memcached.yaml; rm -f config/samples/cache_v1alpha1_memcached.yaml.bak

  header_text "adding policy rbac roles ..."
  sed -i".bak" -E -e '/kubebuilder:scaffold:rules/d' config/rbac/role.yaml; rm -f config/rbac/role.yaml.bak
  cat ../.generate/policy-role.yaml >> config/rbac/role.yaml

  header_text "enabling prometheus metrics..."
  sed -i".bak" -E -e 's/(#- ..\/prometheus)/- ..\/prometheus/g' config/default/kustomization.yaml; rm -f config/default/kustomization.yaml.bak

  header_text "bulding the project ..."
  make docker-build IMG=$operIMG

  header_text "integrating with OLM ..."
  sed -i".bak" -E -e 's/operator-sdk generate kustomize manifests/operator-sdk generate kustomize manifests --interactive=false/g' Makefile; rm -f Makefile.bak

  header_text "generating bundle and building the image $bundleIMG ..."
  make bundle IMG=$operIMG
  make bundle-build BUNDLE_IMG=$bundleIMG
}

gen_helm_sample
