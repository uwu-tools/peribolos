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

package options

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/caarlos0/env/v6"
	actions "github.com/sethvargo/go-githubactions"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/flagutil"
)

const (
	defaultMinAdmins = 5
	defaultDelta     = 0.25
	defaultTokens    = 300
	defaultBurst     = 100
)

type Options struct {
	// Infer if peribolos is running in a GitHub Action.
	// The `CI` environment variable will always be set to "true" in a GitHub Action.
	// ref: https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
	UsingActions      bool `env:"CI"`
	Config            string
	Confirm           bool
	Dump              string
	DumpFull          bool
	MaxDelta          float64
	MinAdmins         int
	RequireSelf       bool
	RequiredAdmins    []string
	FixOrg            bool
	FixOrgMembers     bool
	FixTeamMembers    bool
	FixTeams          bool
	FixTeamRepos      bool
	FixRepos          bool
	IgnoreSecretTeams bool
	AllowRepoArchival bool
	AllowRepoPublish  bool
	GithubOpts        flagutil.GitHubOptions

	// TODO(petr-muller): Remove after August 2021, replaced by github.ThrottleHourlyTokens
	tokenBurst    int
	tokensPerHour int

	logLevel string
}

func New() Options {
	o := new()
	/*
		if err := o.parseArgs(flag.CommandLine, os.Args[1:]); err != nil {
			logrus.Fatalf("Invalid flags: %v", err)
		}
	*/

	return *o
}

func new() *Options {
	o := &Options{}
	if err := env.Parse(o); err != nil {
		fmt.Printf("could not parse env vars, using default options: %v", err)
	}

	return o
}

func (o *Options) validateArgsForAction() error {
	if err := o.GithubOpts.Validate(!o.Confirm); err != nil {
		return fmt.Errorf(err.Error())
	}

	if o.MinAdmins < 2 {
		return fmt.Errorf("--min-admins=%d must be at least 2", o.MinAdmins)
	}

	if o.MaxDelta > 1 || o.MaxDelta < 0 {
		return fmt.Errorf("--maximum-removal-delta=%f must be a non-negative number less than 1.0", o.MaxDelta)
	}

	if o.Confirm && o.Dump != "" {
		return fmt.Errorf("--confirm cannot be used with --dump=%s", o.Dump)
	}
	if o.Config == "" && o.Dump == "" {
		return fmt.Errorf("--config-path or --dump required")
	}
	if o.Config != "" && o.Dump != "" {
		return fmt.Errorf("--config-path=%s and --dump=%s cannot both be set", o.Config, o.Dump)
	}

	if o.DumpFull && o.Dump == "" {
		return fmt.Errorf("--dump-full can't be used without --dump")
	}

	if o.FixTeamMembers && !o.FixTeams {
		return fmt.Errorf("--fix-team-members requires --fix-teams")
	}

	if o.FixTeamRepos && !o.FixTeams {
		return fmt.Errorf("--fix-team-repos requires --fix-teams")
	}

	level, err := logrus.ParseLevel(o.logLevel)
	if err != nil {
		return fmt.Errorf("--log-level invalid: %s", err.Error())
	}
	logrus.SetLevel(level)

	return nil
}

func (o *Options) ParseFromAction() error {
	ghFlags := flag.NewFlagSet("github-flags", flag.ContinueOnError)
	o.GithubOpts.AddCustomizedFlags(ghFlags, flagutil.ThrottlerDefaults(defaultTokens, defaultBurst))

	o.Dump = actions.GetInput(flagDump)
	o.Config = actions.GetInput(flagConfigPath)

	dumpFull := actions.GetInput(flagDumpFull)
	if dumpFull != "" {
		o.DumpFull, _ = strconv.ParseBool(dumpFull)
	}

	confirm := actions.GetInput(flagConfirm)
	if confirm != "" {
		o.Confirm, _ = strconv.ParseBool(confirm)
	}

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.TokenPath = actions.GetInput("github-token-path")
	if o.GithubOpts.TokenPath == "" {
		return fmt.Errorf("missing 'github-token-path'")
	}

	FixOrg := actions.GetInput(flagFixOrg)
	if FixOrg != "" {
		o.FixOrg, _ = strconv.ParseBool(FixOrg)
	}

	FixOrgMembers := actions.GetInput(flagFixOrgMembers)
	if FixOrgMembers != "" {
		o.FixOrgMembers, _ = strconv.ParseBool(FixOrgMembers)
	}

	FixTeams := actions.GetInput(flagFixTeams)
	if FixTeams != "" {
		o.FixTeams, _ = strconv.ParseBool(FixTeams)
	}

	FixTeamMembers := actions.GetInput(flagFixTeamMembers)
	if FixTeamMembers != "" {
		o.FixTeamMembers, _ = strconv.ParseBool(FixTeamMembers)
	}

	FixTeamRepos := actions.GetInput(flagFixTeamRepos)
	if FixTeamRepos != "" {
		o.FixTeamRepos, _ = strconv.ParseBool(FixTeamRepos)
	}

	FixRepos := actions.GetInput(flagFixRepos)
	if FixRepos != "" {
		o.FixRepos, _ = strconv.ParseBool(FixRepos)
	}

	o.MinAdmins = defaultMinAdmins
	MinAdmins := actions.GetInput(flagMinAdmins)
	if MinAdmins != "" {
		o.MinAdmins, _ = strconv.Atoi(MinAdmins)
	}

	RequireSelf := actions.GetInput(flagRequireSelf)
	if RequireSelf != "" {
		o.RequireSelf, _ = strconv.ParseBool(RequireSelf)
	}

	throttleHourlyTokens := actions.GetInput("github-hourly-tokens")
	if throttleHourlyTokens != "" {
		o.GithubOpts.ThrottleHourlyTokens, _ = strconv.Atoi(throttleHourlyTokens)
	}

	throttleAllowBurst := actions.GetInput("github-allowed-burst")
	if throttleHourlyTokens != "" {
		o.GithubOpts.ThrottleAllowBurst, _ = strconv.Atoi(throttleAllowBurst)
	}

	o.logLevel = logrus.InfoLevel.String()
	logLevel := actions.GetInput(flagLogLevel)
	if logLevel != "" {
		o.logLevel = logLevel
	}

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.ThrottleHourlyTokens = o.tokensPerHour

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.ThrottleAllowBurst = o.tokenBurst

	return o.validateArgsForAction()
}
