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

package root

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v7"
	actions "github.com/sethvargo/go-githubactions"
	"github.com/sirupsen/logrus"
	"github.com/uwu-tools/peribolos/third_party/k8s.io/test-infra/prow/flagutil"
)

const (
	defaultMinAdmins = 5
	defaultDelta     = 0.25
	defaultTokens    = 300
	defaultBurst     = 100
)

type Options struct {
	// Configuration settings.

	// Infer if peribolos is running in a GitHub Action.
	// The `CI` environment variable will always be set to "true" in a GitHub Action.
	// ref: https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
	UsingActions bool `env:"CI"`
	Config       string
	Confirm      bool
	Dump         string
	DumpFull     bool
	logLevel     string

	// Protections.
	MaxDelta       float64
	MinAdmins      int
	RequireSelf    bool
	RequiredAdmins []string

	// Organization settings.
	FixOrg         bool
	IgnoreInvitees bool

	// Members settings.
	FixOrgMembers bool

	// Team settings.
	FixTeams          bool
	FixTeamMembers    bool
	FixTeamRepos      bool
	IgnoreSecretTeams bool

	// Repo settings.
	FixRepos          bool
	AllowRepoArchival bool
	AllowRepoPublish  bool

	// Prow GitHub settings.
	GithubOpts flagutil.GitHubOptions
}

func NewOptions() Options {
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
		return err
	}

	if o.MinAdmins < 2 {
		return fmt.Errorf("--min-admins=%d must be at least 2", o.MinAdmins)
	}

	if o.MaxDelta > 1 || o.MaxDelta < 0 {
		return fmt.Errorf("--maximum-removal-delta=%f must be a non-negative number less than 1.0", o.MaxDelta)
	}

	if o.Confirm && o.Dump != "" && o.GithubOpts.AppID == "" {
		return fmt.Errorf("--confirm cannot be used with --dump=%s", o.Dump)
	}

	if o.Dump != "" && !o.Confirm && o.GithubOpts.AppID != "" {
		return fmt.Errorf("--confirm has to be used with --dump=%s and --github-app-id", o.Dump)
	}

	if o.Config == "" && o.Dump == "" {
		return errors.New("--config-path or --dump required")
	}

	if o.Config != "" && o.Dump != "" {
		return fmt.Errorf("--config-path=%s and --dump=%s cannot both be set", o.Config, o.Dump)
	}

	if o.DumpFull && o.Dump == "" {
		return errors.New("--dump-full can't be used without --dump")
	}

	if o.FixTeamMembers && !o.FixTeams {
		return errors.New("--fix-team-members requires --fix-teams")
	}

	if o.FixTeamRepos && !o.FixTeams {
		return errors.New("--fix-team-repos requires --fix-teams")
	}

	level, err := logrus.ParseLevel(o.logLevel)
	if err != nil {
		return fmt.Errorf("--log-level invalid: %s", err.Error())
	}
	logrus.SetLevel(level)

	return nil
}

func (o *Options) ParseFromAction() error {
	// Configuration settings.
	o.Config = actions.GetInput(flagConfigPath)

	confirm := actions.GetInput(flagConfirm)
	if confirm != "" {
		o.Confirm, _ = strconv.ParseBool(confirm)
	}

	o.Dump = actions.GetInput(flagDump)

	dumpFull := actions.GetInput(flagDumpFull)
	if dumpFull != "" {
		o.DumpFull, _ = strconv.ParseBool(dumpFull)
	}

	o.logLevel = logrus.InfoLevel.String()
	logLevel := actions.GetInput(flagLogLevel)
	if logLevel != "" {
		o.logLevel = logLevel
	}

	// Protections.
	o.MaxDelta = defaultDelta
	maxDelta := actions.GetInput(flagMaxRemovalDelta)
	if maxDelta != "" {
		o.MaxDelta, _ = strconv.ParseFloat(maxDelta, 64)
	}

	o.MinAdmins = defaultMinAdmins
	minAdmins := actions.GetInput(flagMinAdmins)
	if minAdmins != "" {
		o.MinAdmins, _ = strconv.Atoi(minAdmins)
	}

	requireSelf := actions.GetInput(flagRequireSelf)
	if requireSelf != "" {
		o.RequireSelf, _ = strconv.ParseBool(requireSelf)
	}

	requiredAdmins := actions.GetInput(flagRequiredAdmins)
	if requiredAdmins != "" {
		// TODO(options): Test this with unexpected inputs as well, including spaces between commas
		o.RequiredAdmins = strings.Split(requiredAdmins, ",")
	}

	// Organization settings.
	fixOrg := actions.GetInput(flagFixOrg)
	if fixOrg != "" {
		o.FixOrg, _ = strconv.ParseBool(fixOrg)
	}

	ignoreInvitees := actions.GetInput(flagIgnoreInvitees)
	if ignoreInvitees != "" {
		o.IgnoreInvitees, _ = strconv.ParseBool(ignoreInvitees)
	}

	// Members settings.
	fixOrgMembers := actions.GetInput(flagFixOrgMembers)
	if fixOrgMembers != "" {
		o.FixOrgMembers, _ = strconv.ParseBool(fixOrgMembers)
	}

	// Team settings.
	fixTeams := actions.GetInput(flagFixTeams)
	if fixTeams != "" {
		o.FixTeams, _ = strconv.ParseBool(fixTeams)
	}

	fixTeamMembers := actions.GetInput(flagFixTeamMembers)
	if fixTeamMembers != "" {
		o.FixTeamMembers, _ = strconv.ParseBool(fixTeamMembers)
	}

	fixTeamRepos := actions.GetInput(flagFixTeamRepos)
	if fixTeamRepos != "" {
		o.FixTeamRepos, _ = strconv.ParseBool(fixTeamRepos)
	}

	ignoreSecretTeams := actions.GetInput(flagIgnoreSecretTeams)
	if ignoreSecretTeams != "" {
		o.IgnoreSecretTeams, _ = strconv.ParseBool(ignoreSecretTeams)
	}

	// Repo settings.
	fixRepos := actions.GetInput(flagFixRepos)
	if fixRepos != "" {
		o.FixRepos, _ = strconv.ParseBool(fixRepos)
	}

	allowRepoArchival := actions.GetInput(flagAllowRepoArchival)
	if allowRepoArchival != "" {
		o.AllowRepoArchival, _ = strconv.ParseBool(allowRepoArchival)
	}

	allowRepoPublish := actions.GetInput(flagAllowRepoPublish)
	if allowRepoPublish != "" {
		o.AllowRepoPublish, _ = strconv.ParseBool(allowRepoPublish)
	}

	// Prow GitHub settings.
	ghFlags := flag.NewFlagSet("github-flags", flag.ContinueOnError)
	o.GithubOpts.AddCustomizedFlags(ghFlags, flagutil.ThrottlerDefaults(defaultTokens, defaultBurst))

	// TODO(flags): Consider parameterizing flag.
	o.GithubOpts.TokenPath = actions.GetInput("github-token-path")
	if o.GithubOpts.TokenPath == "" {
		return fmt.Errorf("missing 'github-token-path'")
	}

	// TODO(action): Is this actually required for GitHub Actions?
	throttleHourlyTokens := actions.GetInput("github-hourly-tokens")
	if throttleHourlyTokens != "" {
		o.GithubOpts.ThrottleHourlyTokens, _ = strconv.Atoi(throttleHourlyTokens)
	}

	// TODO(action): Is this actually required for GitHub Actions?
	throttleAllowBurst := actions.GetInput("github-allowed-burst")
	if throttleHourlyTokens != "" {
		o.GithubOpts.ThrottleAllowBurst, _ = strconv.Atoi(throttleAllowBurst)
	}

	return o.validateArgsForAction()
}
