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

RELEASE_VERSION=v1.1.0

function install_bin() {
  OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(shell uname -m | sed 's/x86_64/amd64/')
	local url="https://github.com/operator-framework/operator-sdk/releases/download/${2}/${1}-${2}-${ARCH}-${OS}"
	curl -sSLo $1 $url
	chmod +x $1
	...
}

install_bin operator-sdk $RELEASE_VERSION
install_bin helm-operator $RELEASE_VERSION
install_bin ansible-operator $RELEASE_VERSION

ROOTDIR="$(pwd)"
cd ../go/.generate/
./gen-go-sample.sh

cd $ROOTDIR
cd ../helm/.generate/
./gen-helm-memcached.sh

cd $ROOTDIR
cd ../ansible/.generate/
./gen-ansible-memcached.sh
