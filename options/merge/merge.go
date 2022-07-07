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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/config/org"
)

type Options struct {
	Orgs        flagMap
	MergeTeams  bool
	IgnoreTeams bool
}

func NewOptions() *Options {
	o := &Options{
		Orgs: flagMap{},
	}

	return o
}

var (
	errValidate = errors.New("some options could not be validated")
)

// Run merges org configuration files
func (o *Options) Run() error {
	cfg, err := loadOrgs(*o)
	if err != nil {
		return fmt.Errorf("Failed to load orgs: %v", err)
	}

	pc := org.FullConfig{
		Orgs: cfg,
	}
	out, err := yaml.Marshal(pc)
	if err != nil {
		return fmt.Errorf("Failed to marshal orgs: %v", err)
	}

	fmt.Println(string(out))
	return nil
}

// Validate validates merge options.
// TODO(options): Cleanup error messages.
func (o *Options) Validate() error {
	var errs []error

	if o.MergeTeams && o.IgnoreTeams {
		errs = append(errs, errors.New("--merge-teams XOR --ignore-teams, not both"))
	}

	if len(errs) != 0 {
		return fmt.Errorf(
			"%w: %+v",
			errValidate,
			errs,
		)
	}

	return nil
}

func unmarshal(path string) (*org.Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read: %v", err)
	}
	var cfg org.Config
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}
	return &cfg, nil
}

func loadOrgs(o Options) (map[string]org.Config, error) {
	config := map[string]org.Config{}
	for name, path := range o.Orgs {
		cfg, err := unmarshal(path)
		if err != nil {
			return nil, fmt.Errorf("error in %s: %v", path, err)
		}
		switch {
		case o.IgnoreTeams:
			cfg.Teams = nil
		case o.MergeTeams:
			if cfg.Teams == nil {
				cfg.Teams = map[string]org.Team{}
			}
			prefix := filepath.Dir(path)
			err := filepath.Walk(prefix, func(path string, info os.FileInfo, err error) error {
				switch {
				case path == prefix:
					return nil // Skip base dir
				case info.IsDir() && filepath.Dir(path) != prefix:
					logrus.Infof("Skipping %s and its children", path)
					return filepath.SkipDir // Skip prefix/foo/bar/ dirs
				case !info.IsDir() && filepath.Dir(path) == prefix:
					return nil // Ignore prefix/foo files
				case filepath.Base(path) == "teams.yaml":
					teamCfg, err := unmarshal(path)
					if err != nil {
						return fmt.Errorf("error in %s: %v", path, err)
					}

					for name, team := range teamCfg.Teams {
						cfg.Teams[name] = team
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("merge teams %s: %v", path, err)
			}
		}
		config[name] = *cfg
	}
	return config, nil
}
