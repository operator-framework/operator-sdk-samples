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

#===================================================================
# FUNCTION trap_add ()
#
# Purpose:  prepends a command to a trap
#
# - 1st arg:  code to add
# - remaining args:  names of traps to modify
#
# Example:  trap_add 'echo "in trap DEBUG"' DEBUG
#
# See: http://stackoverflow.com/questions/3338030/multiple-bash-traps-for-the-same-signal
#===================================================================

function trap_add() {
    trap_add_cmd=$1; shift || fatal "${FUNCNAME} usage error"
    new_cmd=
    for trap_add_name in "$@"; do
        # Grab the currently defined trap commands for this trap
        existing_cmd=`trap -p "${trap_add_name}" |  awk -F"'" '{print $2}'`        # Define default command
        [ -z "${existing_cmd}" ] && existing_cmd="echo exiting @ `date`"        # Generate the new command
        new_cmd="${trap_add_cmd};${existing_cmd}"        # Assign the test
         trap   "${new_cmd}" "${trap_add_name}" || \
                fatal "unable to add to trap ${trap_add_name}"
    done
}
function header_text {
  echo "$header$*$reset"
}
# edit the release version before execute the script
RELEASE_VERSION=v1.1.0
function install_bin() {
  header_text "installing the SDK version ${RELEASE_VERSION}"
  TMPDIR="$(mktemp -d)"
  trap_add 'rm -rf $TMPDIR' EXIT
  pushd "$TMPDIR"
  git clone https://github.com/operator-framework/operator-sdk
  cd operator-sdk
  git checkout ${RELEASE_VERSION}
  git checkout -b generate-release-${RELEASE_VERSION}
  make tidy
  make install
  operator-sdk version
}

ROOTDIR="$(pwd)"

install_bin

header_text "updating samples"

cd $ROOTDIR
cd ../go/.generate/
./gen-go-sample.sh

cd $ROOTDIR
cd ../helm/.generate/
./gen-helm-memcached.sh

cd $ROOTDIR
cd ../ansible/.generate/
./gen-ansible-memcached.sh
