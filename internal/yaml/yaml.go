// Copyright 2023 uwu-tools Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package yaml

import "sigs.k8s.io/yaml"

// Marshal marshals the object into JSON then converts JSON to YAML and returns the
// YAML.
func Marshal(o interface{}) ([]byte, error) {
	return yaml.Marshal(o)
}

// Unmarshal strictly converts YAML to JSON then uses JSON to unmarshal
// into an object, optionally configuring the behavior of the JSON unmarshal.
func Unmarshal(y []byte, o interface{}, opts ...yaml.JSONOpt) error {
	return yaml.UnmarshalStrict(y, o, opts...)
}
