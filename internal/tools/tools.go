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

//go:build tools
// +build tools

// Package tools is used to track binary dependencies with go modules
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	// TODO(tools): Review and enable these dependencies
	//_ "github.com/maxbrunsfeld/counterfeiter/v6"
	//_ "github.com/psampaz/go-mod-outdated"
	//_ "sigs.k8s.io/zeitgeist"

	_ "github.com/airconduct/go-probot"
)
