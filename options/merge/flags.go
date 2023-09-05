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

package merge

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AddFlags adds this options' flags to the cobra command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().Var(
		&o.Orgs,
		"org-part",
		"Each instance adds an org-name=org.yaml part",
	)

	cmd.Flags().BoolVar(
		&o.MergeTeams,
		"merge-teams",
		false,
		"Merge team-name/team.yaml files in each org.yaml dir",
	)

	cmd.Flags().BoolVar(
		&o.IgnoreTeams,
		"ignore-teams",
		false,
		"Never configure teams",
	)

	for _, a := range cmd.Flags().Args() {
		logrus.Print("Extra", a)
		_ = o.Orgs.Set(a)
	}
}
