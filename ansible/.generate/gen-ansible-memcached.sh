#!/usr/bin/env bash

set -o errexit
set -o pipefail

# NO_COLOR=1 # Uncomment to disable color.
NO_COLOR=${NO_COLOR:-""}
if [ -z "$NO_COLOR" ]; then
  header=$'\e[1;33m'
  reset=$'\e[0m'
else
  header=''
  reset=''
fi

operatorName="memcached-operator"
operIMG="quay.io/example-inc/memcached-operator:v0.0.1"
bundleIMG="quay.io/example-inc/memcached-operator-bundle:v0.0.1"

function header_text {
  echo "$header$*$reset"
}

function gen_ansible_memcached {
  header_text "Regenerating Ansible Memcached sample."

  header_text "Removing previous sample."
  cd ..
  mv $operatorName/README.md .generate/README-BAK.md
  rm -rf $operatorName

  header_text "creating $operatorName"
  mkdir $operatorName
  cd $operatorName
  operator-sdk init --plugins=ansible --group=cache --domain=example.com --version=v1alpha1 --kind=Memcached --generate-role


  header_text "Customizing for Memcached"
  cd ../.generate

  header_text "... adding README"
  mv README-BAK.md ../$operatorName/README.md

  header_text "... adding Ansible task and variable"
  cat role_main.yml >> ../$operatorName/roles/memcached/tasks/main.yml
  cat defaults_main.yml >> ../$operatorName/roles/memcached/defaults/main.yml
  sed -i 's|foo: bar|size: 1|g' ../$operatorName/config/samples/cache_v1alpha1_memcached.yaml

  header_text "... adding molecule test for Ansible task"
  sed -i 's/false/true/' ../$operatorName/molecule/default/tasks/memcached_test.yml
  cp size_podcount_test.yml ../$operatorName/molecule/default/tasks/size_podcount_test.yml

  header_text "bulding the project ..."
  make docker-build IMG=$operIMG

  header_text "integrating with OLM ..."
  sed -i".bak" -E -e 's/operator-sdk generate kustomize manifests/operator-sdk generate kustomize manifests --interactive=false/g' Makefile; rm -f Makefile.bak

  header_text "generating bundle and building the image $bundleIMG ..."
  make bundle IMG=$operIMG
  make bundle-build BUNDLE_IMG=$bundleIMG
}

gen_ansible_memcached
