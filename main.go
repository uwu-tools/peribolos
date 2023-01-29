/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/logrusutil"

	"github.com/uwu-tools/peribolos/cmd"
	"github.com/uwu-tools/peribolos/options/root"
)

func main() {
	logrusutil.ComponentInit()

	o := root.NewOptions()
	if o.UsingActions {
		fmt.Println(">>> Running in GitHub Actions environment <<<")
		err := o.ParseFromAction()
		if err != nil {
			fmt.Println(err.Error())
			logrus.WithError(err).Fatal("an error occurred while running peribolos in GitHub Action mode")
		}
	}
	if err := cmd.New(&o).Execute(); err != nil {
		logrus.WithError(err).Fatal("an error occurred while running peribolos")
	}
}
