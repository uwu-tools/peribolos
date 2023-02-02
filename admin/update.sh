#!/usr/bin/env bash
# Copyright 2018 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail
set -x

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
readonly REPO_ROOT

readonly admins=(
  auggie-bot
  cpanato
  justaugustus
)

readonly min_admins="3"

# TODO(merge): Consider making config path configurable
readonly config_path="config"

cd "${REPO_ROOT}"
make update-prep
cmd="${REPO_ROOT}/_output/bin/peribolos"
args=(
  --config-path="$config_path"
  --fix-org
  --fix-org-members
  --fix-teams
  --fix-team-members
  --min-admins="$min_admins"
  "${admins[@]/#/--required-admins=}"
)

"${cmd}" "${args[@]}" "${@}"
