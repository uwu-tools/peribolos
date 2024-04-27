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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	proworg "sigs.k8s.io/prow/pkg/config/org"
	"sigs.k8s.io/release-utils/version"

	"github.com/uwu-tools/peribolos/internal/yaml"
	"github.com/uwu-tools/peribolos/options/merge"
	"github.com/uwu-tools/peribolos/options/root"
	"github.com/uwu-tools/peribolos/org"
)

// New creates a new instance of the peribolos command.
func New(o *root.Options) *cobra.Command {
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

	if !o.UsingActions {
		o.AddFlags(cmd)
	}

	// Add sub-commands.
	cmd.AddCommand(Merge())
	cmd.AddCommand(version.Version())

	return cmd
}

func rootCmd(o *root.Options) error {
	githubClient, err := o.GithubOpts.GitHubClient(!o.Confirm)
	if err != nil {
		logrus.WithError(err).Fatal("Error getting GitHub client.")
	}

	if o.Dump != "" {
		ret, err := org.Dump(githubClient, o.Dump, o.IgnoreSecretTeams, o.GithubOpts.AppID)
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

	// Check if the config path exists
	fileInfo, err := os.Stat(o.Config)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not retrieve file info for %s", o.Config)
	}

	var cfg proworg.FullConfig
	var raw []byte
	if fileInfo.IsDir() {
		files, err := os.ReadDir(o.Config)
		if err != nil {
			logrus.WithError(err).Fatalf("Could not %s directory", o.Config)
		}

		mergeOpts := merge.NewOptions()
		mergeOpts.MergeTeams = true
		configFileName := "org.yaml"
		for _, f := range files {
			if f.IsDir() {
				orgName := f.Name()
				configPath := filepath.Join(o.Config, orgName, configFileName)

				logrus.Infof("Adding config for org: %s", orgName)
				mergeOpts.Orgs[orgName] = configPath
			}
		}

		mergedConfig, err := mergeOpts.Run()
		if err != nil {
			logrus.WithError(err).Fatal("Merging org configs")
		}

		cfg = *mergedConfig
	} else {
		raw, err = os.ReadFile(o.Config)
		if err != nil {
			logrus.WithError(err).Fatal("Could not read --config-path file")
		}

		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			logrus.WithError(err).Fatal("Failed to load configuration")
		}
	}

	for name, orgcfg := range cfg.Orgs {
		if err := org.Configure(*o, githubClient, name, orgcfg); err != nil {
			logrus.Fatalf("Configuration failed: %v", err)
		}
	}

	logrus.Info("Finished syncing configuration.")

	return nil
}
