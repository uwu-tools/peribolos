/*
Copyright RelEngFam Authors.

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

package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	proworg "k8s.io/test-infra/prow/config/org"
	"sigs.k8s.io/release-utils/version"
	"sigs.k8s.io/yaml"

	"github.com/relengfam/peribolos/options"
	"github.com/relengfam/peribolos/org"
)

// New creates a new instance of the peribolos command.
func New(o *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		// TODO(cmd): Add peribolos usage
		Use:   "",
		Short: "",
		Long:  "",
		// TODO(cmd): Add PreRunE logic
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd(o)
		},
	}

	// TODO(cmd): Add flags
	//o.AddFlags(cmd)

	// Add sub-commands.
	cmd.AddCommand(version.Version())
	return cmd
}

func rootCmd(o *options.Options) error {
	githubClient, err := o.GithubOpts.GitHubClient(!o.Confirm)
	if err != nil {
		logrus.WithError(err).Fatal("Error getting GitHub client.")
	}

	if o.Dump != "" {
		ret, err := org.Dump(githubClient, o.Dump, o.IgnoreSecretTeams)
		if err != nil {
			logrus.WithError(err).Fatalf("Dump %s failed to collect current data.", o.Dump)
		}
		var output interface{}
		if o.DumpFull {
			output = proworg.FullConfig{
				Orgs: map[string]proworg.Config{o.Dump: *ret},
			}
		} else {
			output = ret
		}
		out, err := yaml.Marshal(output)
		if err != nil {
			logrus.WithError(err).Fatalf("Dump %s failed to marshal output.", o.Dump)
		}

		logrus.Infof("Dumping orgs[\"%s\"]:", o.Dump)
		fmt.Println(string(out))

		return nil
	}

	raw, err := ioutil.ReadFile(o.Config)
	if err != nil {
		logrus.WithError(err).Fatal("Could not read --config-path file")
	}

	var cfg proworg.FullConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	for name, orgcfg := range cfg.Orgs {
		if err := org.Configure(*o, githubClient, name, orgcfg); err != nil {
			logrus.Fatalf("Configuration failed: %v", err)
		}
	}

	logrus.Info("Finished syncing configuration.")

	return nil
}
