#!/usr/bin/env bash
# Copyright 2022 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIR="$(dirname "${BASH_SOURCE[0]}")"

DIR="$(realpath "${DIR}")"

RELEASES=()

function usage() {
  echo "Usage: $0 <kube-version...>"
  echo "  <kube-version> is the version of kubernetes to test against."
}

function args() {
  if [[ $# -eq 0 ]]; then
    usage
    exit 1
  fi
  while [[ $# -gt 0 ]]; do
    RELEASES+=("${1}")
    shift
  done
}

function test_create_cluster() {
  local release="${1}"
  local name="${2}"

  KWOK_KUBE_VERSION="${release}" kwokctl create cluster --name "${name}" --quiet-pull --kube-audit-policy="${DIR}/audit-policy.yaml"
  if [[ $? -ne 0 ]]; then
    echo "Error: Cluster ${name} creation failed"
    exit 1
  fi
}

function test_delete_cluster() {
  local release="${1}"
  local name="${2}"
  kwokctl delete cluster --name "${name}"
}

function test_audit() {
  local release="${1}"
  local name="${2}"
  local logs
  logs="$(kwokctl logs --name "${name}" audit)"
  if [[ "${logs}" == "" ]]; then
    echo "Error: Audit log is empty"
    return 1
  fi
  echo "${logs}" | head -n 100
}

function main() {
  local failed=()
  for release in "${RELEASES[@]}"; do
    echo "------------------------------"
    echo "Testing audit on ${KWOK_RUNTIME} for ${release}"
    name="audit-cluster-${KWOK_RUNTIME}-${release//./-}"
    test_create_cluster "${release}" "${name}" || failed+=("create_cluster_${name}")
    test_audit "${release}" "${name}" || failed+=("audit_${name}")
    test_delete_cluster "${release}" "${name}" || failed+=("delete_cluster_${name}")
  done

  if [[ "${#failed[@]}" -ne 0 ]]; then
    echo "------------------------------"
    echo "Error: Some tests failed"
    for test in "${failed[@]}"; do
      echo " - ${test}"
    done
    exit 1
  fi
}

args "$@"

main
