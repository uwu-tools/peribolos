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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v6"
	actions "github.com/sethvargo/go-githubactions"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
)

const (
	defaultMinAdmins = 5
	defaultDelta     = 0.25
	defaultTokens    = 300
	defaultBurst     = 100

	// Flags.
	flagRequiredAdmins    = "required-admins"
	flagMinAdmins         = "min-admins"
	flagRequireSelf       = "require-self"
	flagMaxRemovalDelta   = "maximum-removal-delta"
	flagConfigPath        = "config-path"
	flagConfirm           = "confirm"
	flagTokens            = "tokens"
	flagTokenBurst        = "token-burst"
	flagDump              = "dump"
	flagDumpFull          = "dump-full"
	flagIgnoreSecretTeams = "ignore-secret-teams"
	flagFixOrg            = "fix-org"
	flagFixOrgMembers     = "fix-org-members"
	flagFixTeams          = "fix-teams"
	flagFixTeamMembers    = "fix-team-members"
	flagFixTeamRepos      = "fix-team-repos"
	flagFixRepos          = "fix-repos"
	flagAllowRepoArchival = "allow-repo-archival"
	flagAllowRepoPublish  = "allow-repo-publish"
	flagLogLevel          = "log-level"
)

type options struct {
	// Infer if peribolos is running in a GitHub Action.
	// The `CI` environment variable will always be set to "true" in a GitHub Action.
	// ref: https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
	usingActions      bool `env:"CI"`
	config            string
	confirm           bool
	dump              string
	dumpFull          bool
	maximumDelta      float64
	minAdmins         int
	requireSelf       bool
	requiredAdmins    flagutil.Strings
	fixOrg            bool
	fixOrgMembers     bool
	fixTeamMembers    bool
	fixTeams          bool
	fixTeamRepos      bool
	fixRepos          bool
	ignoreSecretTeams bool
	allowRepoArchival bool
	allowRepoPublish  bool
	github            flagutil.GitHubOptions

	// TODO(petr-muller): Remove after August 2021, replaced by github.ThrottleHourlyTokens
	tokenBurst    int
	tokensPerHour int

	logLevel string
}

func newOptions() *options {
	o := &options{}
	if err := env.Parse(o); err != nil {
		fmt.Printf("could not parse env vars, using default options: %v", err)
	}

	return o
}

func parseOptions() options {
	o := newOptions()
	if err := o.parseArgs(flag.CommandLine, os.Args[1:]); err != nil {
		logrus.Fatalf("Invalid flags: %v", err)
	}

	return *o
}

func (o *options) parseArgs(flags *flag.FlagSet, args []string) error {
	o.requiredAdmins = flagutil.NewStrings()

	flags.Var(
		&o.requiredAdmins,
		flagRequiredAdmins,
		"Ensure config specifies these users as admins",
	)

	flags.IntVar(
		&o.minAdmins,
		flagMinAdmins,
		defaultMinAdmins,
		"Ensure config specifies at least this many admins",
	)

	flags.BoolVar(
		&o.requireSelf,
		flagRequireSelf,
		true,
		"Ensure github token path user is an admin",
	)

	flags.Float64Var(
		&o.maximumDelta,
		flagMaxRemovalDelta,
		defaultDelta,
		"Fail if config removes more than this fraction of current members",
	)

	flags.StringVar(
		&o.config,
		flagConfigPath,
		"",
		"Path to org config.yaml",
	)

	flags.BoolVar(
		&o.confirm,
		flagConfirm,
		false,
		"Mutate github if set",
	)

	flags.IntVar(
		&o.tokensPerHour,
		flagTokens,
		defaultTokens,
		"Throttle hourly token consumption (0 to disable) DEPRECATED: use --github-hourly-tokens",
	)

	flags.IntVar(
		&o.tokenBurst,
		flagTokenBurst,
		defaultBurst,
		"Allow consuming a subset of hourly tokens in a short burst. DEPRECATED: use --github-allowed-burst",
	)

	flags.StringVar(
		&o.dump,
		flagDump,
		"",
		"Output current config of this org if set",
	)

	flags.BoolVar(
		&o.dumpFull,
		flagDumpFull,
		false,
		"Output current config of the org as a valid input config file instead of a snippet",
	)

	flags.BoolVar(
		&o.ignoreSecretTeams,
		flagIgnoreSecretTeams,
		false,
		"Do not dump or update secret teams if set",
	)

	flags.BoolVar(
		&o.fixOrg,
		flagFixOrg,
		false,
		"Change org metadata if set",
	)

	flags.BoolVar(
		&o.fixOrgMembers,
		flagFixOrgMembers,
		false,
		"Add/remove org members if set",
	)

	flags.BoolVar(
		&o.fixTeams,
		flagFixTeams,
		false,
		"Create/delete/update teams if set",
	)

	flags.BoolVar(
		&o.fixTeamMembers,
		flagFixTeamMembers,
		false,
		"Add/remove team members if set",
	)

	flags.BoolVar(
		&o.fixTeamRepos,
		flagFixTeamRepos,
		false,
		"Add/remove team permissions on repos if set",
	)

	flags.BoolVar(
		&o.fixRepos,
		flagFixRepos,
		false,
		"Create/update repositories if set",
	)

	flags.BoolVar(
		&o.allowRepoArchival,
		flagAllowRepoArchival,
		false,
		"If set, archiving repos is allowed while updating repos",
	)

	flags.BoolVar(
		&o.allowRepoPublish,
		flagAllowRepoPublish,
		false,
		"If set, making private repos public is allowed while updating repos",
	)

	flags.StringVar(
		&o.logLevel,
		flagLogLevel,
		logrus.InfoLevel.String(),
		fmt.Sprintf("Logging level, one of %v", logrus.AllLevels),
	)

	o.github.AddCustomizedFlags(flags, flagutil.ThrottlerDefaults(defaultTokens, defaultBurst))
	if err := flags.Parse(args); err != nil {
		return err
	}

	o.github.Host = github.DefaultHost

	// TODO(flags): Consider parameterizing flag.
	o.github.ThrottleHourlyTokens = defaultTokens

	// TODO(flags): Consider parameterizing flag.
	o.github.ThrottleAllowBurst = defaultBurst

	if o.usingActions {
		fmt.Printf("Running in GitHub Actions environment")
		err := o.parseFromAction()
		if err != nil {
			return fmt.Errorf("parsing from Action: %w", err)
		}
	}

	if err := o.github.Validate(!o.confirm); err != nil {
		return fmt.Errorf(err.Error())
	}

	if o.minAdmins < 2 {
		return fmt.Errorf("--min-admins=%d must be at least 2", o.minAdmins)
	}

	if o.maximumDelta > 1 || o.maximumDelta < 0 {
		return fmt.Errorf("--maximum-removal-delta=%f must be a non-negative number less than 1.0", o.maximumDelta)
	}

	if o.confirm && o.dump != "" {
		return fmt.Errorf("--confirm cannot be used with --dump=%s", o.dump)
	}
	if o.config == "" && o.dump == "" {
		return fmt.Errorf("--config-path or --dump required")
	}
	if o.config != "" && o.dump != "" {
		return fmt.Errorf("--config-path=%s and --dump=%s cannot both be set", o.config, o.dump)
	}

	if o.dumpFull && o.dump == "" {
		return fmt.Errorf("--dump-full can't be used without --dump")
	}

	if o.fixTeamMembers && !o.fixTeams {
		return fmt.Errorf("--fix-team-members requires --fix-teams")
	}

	if o.fixTeamRepos && !o.fixTeams {
		return fmt.Errorf("--fix-team-repos requires --fix-teams")
	}

	level, err := logrus.ParseLevel(o.logLevel)
	if err != nil {
		return fmt.Errorf("--log-level invalid: %s", err.Error())
	}
	logrus.SetLevel(level)

	return nil
}

func (o *options) parseFromAction() error {
	o.dump = actions.GetInput(flagDump)
	o.config = actions.GetInput(flagConfigPath)

	dumpFull := actions.GetInput(flagDumpFull)
	if dumpFull != "" {
		o.dumpFull, _ = strconv.ParseBool(dumpFull)
	}

	confirm := actions.GetInput(flagConfirm)
	if confirm != "" {
		o.confirm, _ = strconv.ParseBool(confirm)
	}

	// TODO(flags): Consider parameterizing flag.
	o.github.TokenPath = actions.GetInput("github-token-path")
	if o.github.TokenPath == "" {
		return fmt.Errorf("missing 'github-token-path'")
	}

	fixOrg := actions.GetInput(flagFixOrg)
	if fixOrg != "" {
		o.fixOrg, _ = strconv.ParseBool(fixOrg)
	}

	fixOrgMembers := actions.GetInput(flagFixOrgMembers)
	if fixOrgMembers != "" {
		o.fixOrgMembers, _ = strconv.ParseBool(fixOrgMembers)
	}

	fixTeams := actions.GetInput(flagFixTeams)
	if fixTeams != "" {
		o.fixTeams, _ = strconv.ParseBool(fixTeams)
	}

	fixTeamMembers := actions.GetInput(flagFixTeamMembers)
	if fixTeamMembers != "" {
		o.fixTeamMembers, _ = strconv.ParseBool(fixTeamMembers)
	}

	fixTeamRepos := actions.GetInput(flagFixTeamRepos)
	if fixTeamRepos != "" {
		o.fixTeamRepos, _ = strconv.ParseBool(fixTeamRepos)
	}

	fixRepos := actions.GetInput(flagFixRepos)
	if fixRepos != "" {
		o.fixRepos, _ = strconv.ParseBool(fixRepos)
	}

	o.minAdmins = defaultMinAdmins
	minAdmins := actions.GetInput(flagMinAdmins)
	if minAdmins != "" {
		o.minAdmins, _ = strconv.Atoi(minAdmins)
	}

	requireSelf := actions.GetInput(flagRequireSelf)
	if requireSelf != "" {
		o.requireSelf, _ = strconv.ParseBool(requireSelf)
	}

	throttleHourlyTokens := actions.GetInput("github-hourly-tokens")
	if throttleHourlyTokens != "" {
		o.github.ThrottleHourlyTokens, _ = strconv.Atoi(throttleHourlyTokens)
	}

	throttleAllowBurst := actions.GetInput("github-allowed-burst")
	if throttleHourlyTokens != "" {
		o.github.ThrottleAllowBurst, _ = strconv.Atoi(throttleAllowBurst)
	}

	logLevel := actions.GetInput(flagLogLevel)
	if logLevel != "" {
		o.logLevel = logLevel
	}

	return nil
}
