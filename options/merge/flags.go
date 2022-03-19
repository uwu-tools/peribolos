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
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/test-infra/prow/config/org"
)

type flagMap map[string]string

func (fm flagMap) String() string {
	var parts []string
	for key, value := range fm {
		if value == "" {
			parts = append(parts, key)
			continue
		}
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ",")
}

func (fm flagMap) Set(s string) error {
	k, v := parseKeyValue(s)
	if _, present := fm[k]; present {
		return fmt.Errorf("duplicate key: %s", k)
	}
	fm[k] = v
	return nil
}

func (fm flagMap) Type() string {
	return "Type() is not implemented"
}

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
		o.Orgs.Set(a)
	}

	// TODO(merge): Move into mergeCmd()
	cfg, err := loadOrgs(*o)
	if err != nil {
		logrus.Fatalf("Failed to load orgs: %v", err)
	}
	pc := org.FullConfig{
		Orgs: cfg,
	}
	out, err := yaml.Marshal(pc)
	if err != nil {
		logrus.Fatalf("Failed to marshal orgs: %v", err)
	}
	fmt.Println(string(out))
}

func parseKeyValue(s string) (string, string) {
	p := strings.SplitN(s, "=", 2)
	if len(p) == 1 {
		return p[0], ""
	}
	return p[0], p[1]
}
